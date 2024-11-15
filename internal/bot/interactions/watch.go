package interactions

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/angelajfisher/meeting-mate/internal/orchestrator"
	"github.com/angelajfisher/meeting-mate/internal/types"
	"github.com/bwmarrin/discordgo"
)

type watchProcess struct {
	MeetingID         string                 // The ID of the Zoom meeting being watched
	GuildID           string                 // The ID of the Discord guild this watch is for
	channelID         string                 // The ID of the channel this watch is for
	session           *discordgo.Session     // The active Discord session used for communication
	silent            bool                   // Whether messages should be sent with the @silent flag
	joinLink          string                 // User-supplied link for others to join the meeting
	showStats         bool                   // Whether meetings stats should be sent at the end of a meeting
	historyLevel      string                 // How many messages to send / delete as meetings start and end
	meetingInProgress bool                   // Whether the meeting is currently ongoing
	meetingMsgContent *discordgo.MessageSend // The data the message should contain
	meetingStatusMsg  *discordgo.Message     // The message sent by the bot
	restartCommand    string                 // The command to restart this exact watch
	o                 orchestrator.Orchestrator
}

func HandleWatch(s *discordgo.Session, i *discordgo.InteractionCreate, o orchestrator.Orchestrator, opts optionMap) {
	var (
		newMeetingID = opts["meeting_id"].StringValue()
		err          error
		responseMsg  struct {
			msg       string
			flags     discordgo.MessageFlags
			terminate bool
		}
		restartCommandBuilder = new(strings.Builder)
	)
	restartCommandBuilder.WriteString("```/watch meeting_id: " + newMeetingID)

	log.Printf("%s: /watch ID %s in %s", i.Member.User, newMeetingID, i.GuildID)

	// Check if the meeting ID is currently being watched
	if o.IsOngoingWatch(i.GuildID, newMeetingID) {
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
			restartCommandBuilder.WriteString(" join_link: " + joinLink)
		} else {
			responseMsg.msg = "Invalid join link provided. Please ensure your URL is correct and starts with \"https://\""
			responseMsg.flags = discordgo.MessageFlagsEphemeral
			responseMsg.terminate = true
		}
	}

	if !responseMsg.terminate {
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
		restartCommandBuilder.WriteString(" silent: false")
	}
	summary := false
	if v, ok := opts["summary"]; ok {
		summary = v.BoolValue()
		restartCommandBuilder.WriteString(" summary: true")
	}
	keepHistory := PARTIAL_HISTORY
	if v, ok := opts["keep_history"]; ok {
		keepHistory = v.StringValue()
		restartCommandBuilder.WriteString(" keep_history: " + keepHistory)
	}
	restartCommandBuilder.WriteString("```")
	watch := watchProcess{
		MeetingID:         newMeetingID,
		GuildID:           i.GuildID,
		session:           s,
		channelID:         i.ChannelID,
		silent:            sendSilently,
		joinLink:          joinLink,
		showStats:         summary,
		historyLevel:      keepHistory,
		meetingInProgress: false,
		meetingMsgContent: &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{{Type: discordgo.EmbedTypeRich,
			Description: "Loading..."}}},
		meetingStatusMsg: nil,
		restartCommand:   restartCommandBuilder.String(),
		o:                o,
	}
	if watch.silent {
		watch.meetingMsgContent.Flags = discordgo.MessageFlagsSuppressNotifications
	}

	watch.listen()
}

func (w *watchProcess) listen() {
	var (
		err      error
		shutdown = false
	)
	for updateData := range w.o.StartWatch(w.GuildID, w.MeetingID) {
		if updateData.EventType == types.BotShutdown {
			shutdown = true
			w.meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch stopped due to bot shutdown." +
				" Please use the following command to restart when available:\n" + w.restartCommand
			w.meetingMsgContent.Components = []discordgo.MessageComponent{}
			break
		} else if updateData.EventType == types.WatchCanceled {
			w.meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch on this meeting was canceled."
			w.meetingMsgContent.Components = []discordgo.MessageComponent{}
			break
		}

		// Remove old meeting message if needed (full history messages will be nil if not in progress)
		if !w.meetingInProgress && w.meetingStatusMsg != nil {
			func() {
				defer func() { w.meetingStatusMsg = nil }()
				if w.historyLevel == PARTIAL_HISTORY {
					channel, chanErr := w.session.Channel(w.channelID)
					if chanErr != nil {
						log.Printf("could not get channel info: %s", chanErr)
						return
					}
					// Keep meeting history if it's been buried by conversation
					if channel.LastMessageID != w.meetingStatusMsg.ID {
						return
					}
				}
				delErr := w.session.ChannelMessageDelete(w.channelID, w.meetingStatusMsg.ID)
				if delErr != nil {
					log.Printf("could not delete previous meeting message: %s", delErr)
				}
			}()
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
			w.meetingMsgContent.Embeds[0].Title = updateData.MeetingName
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

		w.updateMeetingMsg(updateData)
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
	} else if shutdown {
		_, err = w.session.ChannelMessageSendComplex(w.channelID, w.meetingMsgContent)
		if err != nil {
			log.Printf("could not respond to interaction: %s", err)
		}
	}
}

func (w *watchProcess) updateMeetingMsg(updateData types.UpdateData) {
	var err error

	if updateData.EventType == types.ZoomMeetingEnd {
		w.meetingMsgContent.Embeds[0].Description = "This meeting ended."
		w.meetingInProgress = false
		w.meetingMsgContent.Components = []discordgo.MessageComponent{}
		if w.showStats {
			w.meetingMsgContent.Embeds[0].Fields = []*discordgo.MessageEmbedField{
				{Name: "Summary", Value: fmt.Sprintf(
					"Total Participants: %v\nDuration: %s",
					updateData.TotalParticipants,
					updateData.MeetingDuration,
				)},
			}
		} else {
			w.meetingMsgContent.Embeds[0].Fields = nil
		}
	} else {
		w.meetingMsgContent.Embeds[0].Fields[0].Value = updateData.Participants
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
		// Since all messages are kept with full history, remove reference to old message so it isn't removed
		if !w.meetingInProgress && w.historyLevel == FULL_HISTORY {
			w.meetingStatusMsg = nil
		}
	} else {
		log.Println("could not update meeting status: meeting message is nil")
	}
}
