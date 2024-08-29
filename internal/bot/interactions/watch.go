package interactions

import (
	"log"
	"strings"

	"github.com/angelajfisher/meeting-mate/internal/types"
	"github.com/bwmarrin/discordgo"
)

var (
	Watchers []chan struct{} // Shutdown communication channels for all active meeting watches

	// map[meetingID]map[guildID]struct{}
	//
	// Stores sets of guild IDs that are currently watching a meeting ID
	watchedMeetings = make(map[string]map[string]struct{})
)

type watchProcess struct {
	session           *discordgo.Session     // The active Discord session used for communication
	meetingID         string                 // The ID of the Zoom meeting being watched
	guildID           string                 // The ID of the Discord guild this watch is for
	channelID         string                 // The ID of the channel this watch is for
	silent            bool                   // Whether messages should be sent with the @silent flag
	shutdownCh        chan struct{}          // Channel listening for shutdown commands
	participantsList  map[string]string      // map[participantID]participantName
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
	if watchingGuilds, exist := watchedMeetings[newMeetingID]; exist {
		// Check whether this guild is already watching it
		if _, thisGuild := watchingGuilds[i.GuildID]; thisGuild {
			responseMsg.msg = "Watch on meeting ID " + newMeetingID +
				" is already ongoing. It will continue indefinitely unless `/cancel`ed."
			responseMsg.flags = discordgo.MessageFlagsEphemeral
			responseMsg.terminate = true
		}
	} else if newMeetingID != "" {
		// If the ID wasn't being watched, add it to the system now
		watchedMeetings[newMeetingID] = make(map[string]struct{})
	} else {
		responseMsg.msg = "Please supply a valid meeting ID to watch."
		responseMsg.flags = discordgo.MessageFlagsEphemeral
		responseMsg.terminate = true
	}

	if !responseMsg.terminate {
		watchedMeetings[newMeetingID][i.GuildID] = struct{}{}
		responseMsg.msg = "Initiating watch on meeting ID " + newMeetingID + "!\nStop at any time with `/cancel`"
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

	shutdownCh := make(chan struct{})
	Watchers = append(Watchers, shutdownCh)
	go func() {
		<-shutdownCh
		types.WatchMeetingCh <- types.Shutdown
	}()

	watch := watchProcess{
		session:           s,
		meetingID:         newMeetingID,
		channelID:         i.ChannelID,
		guildID:           i.GuildID,
		silent:            sendSilently,
		shutdownCh:        shutdownCh,
		participantsList:  make(map[string]string),
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
	types.WatchMeetingCh <- w.meetingID

	for zoomData := range types.MeetingDataCh {
		if zoomData.EventType == types.Canceled {
			w.meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch on this meeting was canceled."
			delete(watchedMeetings[w.meetingID], w.guildID)
			break
		} else if zoomData.EventType == types.Shutdown {
			w.meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch stopped due to bot shutdown." +
				" Please restart with `/watch` when available."
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
			w.meetingMsgContent.Embeds[0].Title = zoomData.MeetingName
			w.meetingMsgContent.Embeds[0].Description = "This meeting is in progress."
			w.meetingMsgContent.Embeds[0].Fields = []*discordgo.MessageEmbedField{{Name: "Current Participants"}}
		}

		w.updateMeetingMsg(zoomData)
	}

	// Update any existing status messages w/ notice that the watch stopped
	if w.meetingStatusMsg != nil {
		w.meetingMsgContent.Embeds[0].Fields = nil
		updatedContent := discordgo.MessageEdit{
			Embeds:  &w.meetingMsgContent.Embeds,
			ID:      w.meetingStatusMsg.ID,
			Channel: w.meetingStatusMsg.ChannelID,
		}
		if _, err = w.session.ChannelMessageEditComplex(&updatedContent); err != nil {
			log.Printf("could not respond to interaction: %s", err)
		}
	}
}

func (w *watchProcess) updateMeetingMsg(zoomData types.EventData) {
	var err error

	switch zoomData.EventType {
	case types.ParticipantJoin:
		w.participantsList[zoomData.ParticipantID] = zoomData.ParticipantName
		w.meetingMsgContent.Embeds[0].Fields[0].Value = w.stringifyParticipants()

	case types.ParticipantLeave:
		delete(w.participantsList, zoomData.ParticipantID)
		w.meetingMsgContent.Embeds[0].Fields[0].Value = w.stringifyParticipants()

	case types.MeetingEnd:
		clear(w.participantsList)
		w.meetingMsgContent.Embeds[0].Description = "This meeting ended."
		w.meetingMsgContent.Embeds[0].Fields = nil
		w.meetingInProgress = false
	}

	if w.meetingStatusMsg != nil {
		updatedContent := discordgo.MessageEdit{
			Embeds:  &w.meetingMsgContent.Embeds,
			ID:      w.meetingStatusMsg.ID,
			Channel: w.meetingStatusMsg.ChannelID,
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

func (w *watchProcess) stringifyParticipants() string {
	builder := new(strings.Builder)
	for _, name := range w.participantsList {
		builder.WriteString(name + "\n")
	}
	if builder.String() == "" {
		builder.WriteString("Unknown")
	}
	return builder.String()
}
