package interactions

import (
	"log"
	"strings"

	"github.com/angelajfisher/meeting-mate/internal/types"
	"github.com/bwmarrin/discordgo"
)

const noWatch = "Nothing to cancel: the watch for this meeting is not active."

func HandleCancel(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {
	var (
		err       error
		meetingID string
	)

	// Check for a meeting ID provided with the command
	if v, ok := opts["meeting_id"]; ok && v.StringValue() != "" {
		meetingID = v.StringValue()
		log.Printf("%s: /cancel ID %s", i.Member.User, meetingID)
	} else {
		log.Printf("%s: /cancel", i.Member.User)
	}

	// ID provided path
	if meetingID != "" || types.MeetingWatches.Exists(i.GuildID, meetingID) {
		var (
			response string
			msgFlags discordgo.MessageFlags
		)

		if types.MeetingWatches.Exists(i.GuildID, meetingID) {
			types.DataListeners.Remove(i.GuildID, meetingID, types.EventData{EventType: types.WatchCanceled})
			types.MeetingWatches.Remove(i.GuildID, meetingID)
			response = "Canceled watch on meeting ID `" + meetingID + "`."
		} else {
			response = noWatch
			msgFlags = discordgo.MessageFlagsEphemeral
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
				Flags:   msgFlags,
			},
		})
		if err != nil {
			log.Printf("could not respond to interaction: %s", err)
		}
		return
	}

	// ID not provided path
	// Format all active meeting watches in this server into selectable options
	guildWatches := []discordgo.SelectMenuOption{}
	for meeting := range types.MeetingWatches.GetMeetings(i.GuildID) {
		guildWatches = append(guildWatches, discordgo.SelectMenuOption{
			Label: meeting,
			Value: meeting,
		})
	}

	if len(guildWatches) == 0 {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Nothing to cancel: there are no active watches in this server.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("could not respond to interaction: %s", err)
		}
		return
	}

	minVals := 1
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Which active meeting watches would you like to cancel?",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						MenuType:    discordgo.StringSelectMenu,
						Options:     guildWatches,
						MinValues:   &minVals,
						MaxValues:   len(guildWatches),
						CustomID:    "meeting_cancel_selection" + i.Interaction.Member.User.ID,
						Placeholder: "Select meeting ID(s)",
					},
				}},
			},
			CustomID: "meeting_cancel_modal_" + i.Interaction.Member.User.ID,
		},
	})
	if err != nil {
		log.Printf("could not respond to interaction: %s", err)
	}
}

func HandleCancelSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()

	var responseMsg string
	if len(data.Values) == 1 {
		for _, dataChannel := range types.DataListeners.GetMeetingListeners(data.Values[0]) {
			dataChannel <- types.EventData{EventType: types.WatchCanceled}
		}
		responseMsg = "Canceled watch on meeting ID `" + data.Values[0] + "`."
	} else {
		builder := new(strings.Builder)
		builder.WriteString("Canceled watch on the following meetings:")
		for _, meetingID := range data.Values {
			types.DataListeners.Remove(i.GuildID, meetingID, types.EventData{EventType: types.WatchCanceled})
			types.MeetingWatches.Remove(i.GuildID, meetingID)
			builder.WriteString("\n- `" + meetingID + "`")
		}
		responseMsg = builder.String()
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: responseMsg,
		},
	})
	if err != nil {
		log.Printf("could not respond to interaction: %s", err)
	}
}
