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
	if v, ok := opts["meeting_id"]; ok && v.StringValue() != "" {
		meetingID = v.StringValue()
		log.Printf("%s: /cancel ID %s", i.Member.User, meetingID)
	} else {
		log.Printf("%s: /cancel", i.Member.User)
	}

	if meetingID != "" || watchedMeetings[meetingID] != nil {
		var (
			response string
			msgFlags discordgo.MessageFlags
		)

		if _, thisGuild := watchedMeetings[meetingID][i.GuildID]; thisGuild {
			types.WatchMeetingCh <- types.Canceled
			delete(watchedMeetings[meetingID], i.GuildID)
			response = "Canceled watch on meeting `" + meetingID + "`."
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

	guildWatches := []discordgo.SelectMenuOption{}
	for meeting := range watchedMeetings {
		if _, thisGuild := watchedMeetings[meeting][i.GuildID]; thisGuild {
			guildWatches = append(guildWatches, discordgo.SelectMenuOption{
				Label: meeting,
				Value: meeting,
			})
		}
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
		// todo: cancel the meeting watch
		responseMsg = "Canceled watch on meeting `" + data.Values[0] + "`."
	} else {
		builder := new(strings.Builder)
		builder.WriteString("Canceled watch on the following meetings:")
		for _, meetingID := range data.Values {
			// todo: cancel the meeting watch
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
