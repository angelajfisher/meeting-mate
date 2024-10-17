package interactions

import (
	"log"
	"net/url"

	"github.com/angelajfisher/meeting-mate/internal/types"
	"github.com/bwmarrin/discordgo"
)

var (
	Watches []watchProcess // All ongoing watch processes
)

type watchProcess struct {
	Meeting           *types.Meeting         // The Zoom meeting being watched
	GuildID           string                 // The ID of the Discord guild this watch is for
	channelID         string                 // The ID of the channel this watch is for
	session           *discordgo.Session     // The active Discord session used for communication
	silent            bool                   // Whether messages should be sent with the @silent flag
	joinLink          string                 // User-supplied link for others to join the meeting
	meetingInProgress bool                   // Whether the meeting is currently ongoing
	meetingMsgContent *discordgo.MessageSend // The data the message should contain
	meetingStatusMsg  *discordgo.Message     // The message sent by the bot
}

func HandleWatch(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {
	var (
		newMeetingID = opts["meeting_id"].StringValue()
		err          error
		responseMsg  struct {
			msg       string
			flags     discordgo.MessageFlags
			terminate bool
		}
	)

	log.Printf("%s: /watch ID %s", i.Member.User, newMeetingID)

	// Check if the meeting ID is currently being watched
	if types.MeetingWatches.Exists(i.GuildID, newMeetingID) {
		responseMsg.msg = "Watch on meeting ID `" + newMeetingID +
			"` is already ongoing. It will continue indefinitely unless `/cancel`ed."
		responseMsg.flags = discordgo.MessageFlagsEphemeral
		responseMsg.terminate = true
	} else if newMeetingID == "" {
		responseMsg.msg = "Please supply a valid meeting ID to watch."
		responseMsg.flags = discordgo.MessageFlagsEphemeral
		responseMsg.terminate = true
	}

	// Check for valid join link if provided
	joinLink := ""
	if v, ok := opts["join_link"]; ok {
		u, parseErr := url.Parse(v.StringValue())
		if parseErr == nil && u.Scheme == "https" {
			joinLink = v.StringValue()
		} else {
			responseMsg.msg = "Invalid join link provided. Please ensure your URL is correct and starts with \"https://\""
			responseMsg.flags = discordgo.MessageFlagsEphemeral
			responseMsg.terminate = true
		}
	}

	if !responseMsg.terminate {
		types.MeetingWatches.Add(i.GuildID, newMeetingID)
		responseMsg.msg = "Initiating watch on meeting ID `" + newMeetingID + "`!\nStop at any time with `/cancel`"
	}

	// Send the interaction response message to the user
	if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: responseMsg.msg,
			Flags:   responseMsg.flags,
		},
	}); err != nil {
		log.Printf("could not respond to interaction: %s", err)
	}
	if responseMsg.terminate {
		return
	}

	// Initialize the new watch process
	sendSilently := true
	if v, ok := opts["silent"]; ok && !v.BoolValue() {
		sendSilently = v.BoolValue()
	}
	watch := watchProcess{
		Meeting:           types.AllMeetings.NewMeeting(newMeetingID),
		GuildID:           i.GuildID,
		session:           s,
		channelID:         i.ChannelID,
		silent:            sendSilently,
		joinLink:          joinLink,
		meetingInProgress: false,
		meetingMsgContent: &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{{Type: discordgo.EmbedTypeRich,
			Description: "Loading..."}}},
		meetingStatusMsg: nil,
	}
	if watch.silent {
		watch.meetingMsgContent.Flags = discordgo.MessageFlagsSuppressNotifications
	}

	watch.listen()
}

func (w *watchProcess) listen() {
	var err error

	meetingID := w.Meeting.GetID()
	for zoomData := range types.DataListeners.Listen(w.GuildID, meetingID) {
		if zoomData.EventType == types.WatchCanceled {
			w.meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch on this meeting was canceled."
			w.meetingMsgContent.Components = []discordgo.MessageComponent{}
			types.MeetingWatches.Remove(w.GuildID, meetingID)
			break
		} else if zoomData.EventType == types.BotShutdown {
			w.meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch stopped due to bot shutdown." +
				" Please restart with `/watch " + meetingID + "` when available."
			w.meetingMsgContent.Components = []discordgo.MessageComponent{}
			break
		}

		if w.meetingStatusMsg == nil {
			w.meetingStatusMsg, err = w.session.ChannelMessageSendComplex(w.channelID, w.meetingMsgContent)
			if err != nil {
				log.Printf("could not respond to interaction: %s", err)
			}
		}

		// If there wasn't a meeting in progress before this data came in, start a new meeting message
		if !w.meetingInProgress {
			w.meetingInProgress = true
			w.meetingMsgContent.Embeds[0].Title = w.Meeting.Name(zoomData.MeetingName)
			w.meetingMsgContent.Embeds[0].Description = "This meeting is in progress."
			w.meetingMsgContent.Embeds[0].Fields = []*discordgo.MessageEmbedField{{Name: "Current Participants"}}
			if w.joinLink != "" {
				w.meetingMsgContent.Components = []discordgo.MessageComponent{
					discordgo.ActionsRow{Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label: "Join",
							URL:   w.joinLink,
							Style: discordgo.LinkButton,
						},
					}},
				}
			}
		}

		w.updateMeetingMsg(zoomData)
	}

	// Update any existing status messages w/ notice that the watch stopped
	if w.meetingStatusMsg != nil {
		w.meetingMsgContent.Embeds[0].Fields = nil
		updatedContent := discordgo.MessageEdit{
			Embeds:     &w.meetingMsgContent.Embeds,
			ID:         w.meetingStatusMsg.ID,
			Channel:    w.meetingStatusMsg.ChannelID,
			Components: &[]discordgo.MessageComponent{},
		}
		if _, err = w.session.ChannelMessageEditComplex(&updatedContent); err != nil {
			log.Printf("could not respond to interaction: %s", err)
		}
	}
}

func (w *watchProcess) updateMeetingMsg(zoomData types.EventData) {
	var err error

	switch zoomData.EventType {
	case types.ZoomParticipantJoin:
		w.Meeting.Participants.Add(zoomData.ParticipantID, zoomData.ParticipantName)
		w.meetingMsgContent.Embeds[0].Fields[0].Value = w.Meeting.Participants.Stringify()

	case types.ZoomParticipantLeave:
		w.Meeting.Participants.Remove(zoomData.ParticipantID)
		w.meetingMsgContent.Embeds[0].Fields[0].Value = w.Meeting.Participants.Stringify()

	case types.ZoomMeetingEnd:
		w.Meeting.Participants.Empty()
		w.meetingMsgContent.Embeds[0].Description = "This meeting ended."
		w.meetingMsgContent.Embeds[0].Fields = nil
		w.meetingInProgress = false
		w.meetingMsgContent.Components = []discordgo.MessageComponent{}
	}

	if w.meetingStatusMsg != nil {
		updatedContent := discordgo.MessageEdit{
			Embeds:     &w.meetingMsgContent.Embeds,
			ID:         w.meetingStatusMsg.ID,
			Channel:    w.meetingStatusMsg.ChannelID,
			Components: &w.meetingMsgContent.Components,
		}
		w.meetingStatusMsg, err = w.session.ChannelMessageEditComplex(&updatedContent)
		if err != nil {
			log.Printf("could not respond to interaction: %s\n", err)
		}
		if !w.meetingInProgress {
			w.meetingStatusMsg = nil
		}
	} else {
		log.Println("could not update meeting status: meeting message is nil")
	}
}
