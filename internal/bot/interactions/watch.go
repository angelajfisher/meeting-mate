package interactions

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/angelajfisher/meeting-mate/internal/db"
	"github.com/angelajfisher/meeting-mate/internal/orchestrator"
	"github.com/angelajfisher/meeting-mate/internal/types"
	"github.com/bwmarrin/discordgo"
)

type watchProcess struct {
	meetingID         string                 // The ID of the Zoom meeting being watched
	guildID           string                 // The ID of the Discord guild this watch is for
	flags             types.FeatureFlags     // The user-configurable options for the watch
	channelID         string                 // The ID of the channel this watch is for
	session           *discordgo.Session     // The active Discord session used for communication
	meetingInProgress bool                   // Whether the meeting is currently ongoing
	meetingMsgContent *discordgo.MessageSend // The data the message should contain
	meetingStatusMsg  *discordgo.Message     // The message sent by the bot
	o                 orchestrator.Orchestrator
}

func HandleWatch(s *discordgo.Session, i *discordgo.InteractionCreate, o orchestrator.Orchestrator, opts optionMap) {
	var (
		newMeetingID = opts[MEETING_OPT].StringValue()
		err          error
		responseMsg  struct {
			msg       string
			flags     discordgo.MessageFlags
			terminate bool
		}
	)

	log.Printf("%s: /watch ID %s in %s", i.Member.User, newMeetingID, i.GuildID)

	// Check if the meeting ID is currently being watched
	if o.IsOngoingWatch(i.GuildID, newMeetingID) {
		responseMsg.msg = "Watch on meeting ID `" + newMeetingID +
			"` is already ongoing. It will continue indefinitely unless `/cancel`ed." +
			"\nIf you'd like to change the settings on this watch, try `/update `" + newMeetingID + "`!"
		responseMsg.flags = discordgo.MessageFlagsEphemeral
		responseMsg.terminate = true
	} else if newMeetingID == "" {
		responseMsg.msg = "Please supply a valid meeting ID to watch."
		responseMsg.flags = discordgo.MessageFlagsEphemeral
		responseMsg.terminate = true
	}

	// Check for valid join link if provided
	if v, ok := opts[LINK_OPT]; ok {
		u, parseErr := url.Parse(v.StringValue())
		if parseErr != nil || u.Scheme != "https" {
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
		log.Printf("HandleWatch: could not respond to interaction: %s", err)
	}
	if responseMsg.terminate {
		return
	}

	// Initialize the new watch process
	watch := watchProcess{
		meetingID:         newMeetingID,
		guildID:           i.GuildID,
		flags:             generateWatchFlags(opts),
		session:           s,
		channelID:         i.ChannelID,
		meetingInProgress: false,
		meetingMsgContent: &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{{Type: discordgo.EmbedTypeRich,
			Description: "Loading..."}}},
		meetingStatusMsg: nil,
		o:                o,
	}
	if watch.flags.Silent {
		watch.meetingMsgContent.Flags = discordgo.MessageFlagsSuppressNotifications
	}

	// Store watch details in the database in order to restore the state upon a restart
	o.Database.SaveWatch(db.WatchData{
		MeetingID: watch.meetingID,
		GuildID:   watch.guildID,
		ChannelID: watch.channelID,
		Options:   watch.flags,
	})
	watch.listen("")
}

// Restores an ongoing watch by initializing a watch process with saved data
func LoadSavedWatch(
	s *discordgo.Session,
	o orchestrator.Orchestrator,
	watchData db.WatchData,
) {
	// Initialize the watch process
	watch := watchProcess{
		meetingID:         watchData.MeetingID,
		guildID:           watchData.GuildID,
		flags:             watchData.Options,
		session:           s,
		channelID:         watchData.ChannelID,
		meetingInProgress: false,
		meetingMsgContent: &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{{
			Type:        discordgo.EmbedTypeRich,
			Description: "Loading...",
		}}},
		meetingStatusMsg: nil,
		o:                o,
	}
	if watch.flags.Silent {
		watch.meetingMsgContent.Flags = discordgo.MessageFlagsSuppressNotifications
	}

	watch.listen(watchData.MeetingTopic)
}

// Starts a goroutine to listen to Zoom meeting changes and update the meeting message accordingly
//
//nolint:gocognit
func (w *watchProcess) listen(meetingTopic string) {
	var (
		err      error
		shutdown = false
	)
	for updateData := range w.o.StartWatch(w.guildID, w.meetingID, meetingTopic) {
		if updateData.EventType == types.SYSTEM_SHUTDOWN {
			messageBuilder := new(strings.Builder)
			messageBuilder.WriteString("**Status Unknown**\nThe watch stopped due to bot shutdown.")
			if w.o.Database.Enabled {
				messageBuilder.WriteString(
					" No action is needed on your part — the watch should automatically resume when Meeting Mate returns.",
				)
			} else {
				messageBuilder.WriteString(
					" Please use the following command to restart when available:\n" + w.flags.RestartCommand,
				)
			}
			shutdown = true
			meetingName := w.o.GetMeetingName(w.meetingID)
			if meetingName == "" {
				meetingName = "Meeting ID: " + w.meetingID
			}
			w.meetingMsgContent.Embeds[0].Title = meetingName
			w.meetingMsgContent.Embeds[0].Description = messageBuilder.String()
			w.meetingMsgContent.Components = []discordgo.MessageComponent{}
			break
		}
		if updateData.EventType == types.WATCH_CANCELED {
			w.meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch on this meeting was canceled."
			w.meetingMsgContent.Components = []discordgo.MessageComponent{}
			break
		}
		if updateData.EventType == types.UPDATE_FLAGS {
			w.flags = updateData.Flags
			if w.flags.Silent {
				w.meetingMsgContent.Flags = discordgo.MessageFlagsSuppressNotifications
			} else {
				w.meetingMsgContent.Flags = 0
			}
			continue
		}

		// Remove old meeting message if needed (full history messages will be nil if not in progress)
		if !w.meetingInProgress && w.meetingStatusMsg != nil {
			func() {
				defer func() { w.meetingStatusMsg = nil }()
				if w.flags.HistoryLevel == types.PARTIAL_HISTORY {
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
				log.Printf("WatchListener-Update: could not respond to interaction: %s", err)
			}
		}

		// If there wasn't a meeting in progress before this data came in, start a new meeting message
		if !w.meetingInProgress {
			w.meetingInProgress = true
			w.meetingMsgContent.Embeds[0].Title = updateData.MeetingName
			w.meetingMsgContent.Embeds[0].Description = "This meeting is in progress."
			w.meetingMsgContent.Embeds[0].Fields = []*discordgo.MessageEmbedField{{Name: "Current Participants"}}
			if w.flags.JoinLink != "" {
				w.meetingMsgContent.Components = []discordgo.MessageComponent{
					discordgo.ActionsRow{Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label: "Join",
							URL:   w.flags.JoinLink,
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
			log.Printf("WatchListener-Stop: could not respond to interaction: %s", err)
		}
	} else if shutdown {
		_, err = w.session.ChannelMessageSendComplex(w.channelID, w.meetingMsgContent)
		if err != nil {
			log.Printf("WatchListener-Shutdown: could not respond to interaction: %s", err)
		}
	}
}

func (w *watchProcess) updateMeetingMsg(updateData types.UpdateData) {
	var err error

	if updateData.EventType == types.ZOOM_MEETING_END {
		w.meetingMsgContent.Embeds[0].Description = "This meeting ended."
		w.meetingInProgress = false
		w.meetingMsgContent.Components = []discordgo.MessageComponent{}
		if w.flags.Summaries {
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
			log.Printf("UpdateMeetingMsg: could not respond to interaction: %s\n", err)
		}
		// Since all messages are kept with full history, remove reference to old message so it isn't removed
		if !w.meetingInProgress && w.flags.HistoryLevel == types.FULL_HISTORY {
			w.meetingStatusMsg = nil
		}
	} else {
		log.Println("could not update meeting status: meeting message is nil")
	}
}
