package interactions

import (
	"log"
	"strings"

	"github.com/angelajfisher/meeting-mate/internal/orchestrator"
	"github.com/bwmarrin/discordgo"
)

const noWatch = "Nothing to cancel: there is no active watch on this meeting."

func HandleCancel(s *discordgo.Session, i *discordgo.InteractionCreate, o orchestrator.Orchestrator, opts optionMap) {
	var (
		err       error
		meetingID string
	)

	// Check for a meeting ID provided with the command
	if v, ok := opts["meeting_id"]; ok && v.StringValue() != "" {
		meetingID = v.StringValue()
		log.Printf("%s in %s: /cancel ID %s", i.Member.User, i.GuildID, meetingID)
	} else {
		log.Printf("%s in %s: /cancel", i.Member.User, i.GuildID)
	}

	// ID provided path
	if meetingID != "" || o.IsOngoingWatch(i.GuildID, meetingID) {
		var (
			response string
			msgFlags discordgo.MessageFlags
		)

		if o.IsOngoingWatch(i.GuildID, meetingID) {
			o.CancelWatch(i.GuildID, meetingID)
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
	for _, meeting := range o.GetGuildMeetings(i.GuildID) {
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

func HandleCancelSelection(s *discordgo.Session, i *discordgo.InteractionCreate, o orchestrator.Orchestrator) {
	data := i.MessageComponentData()

	var responseMsg string
	if len(data.Values) == 1 {
		o.CancelWatch(i.GuildID, data.Values[0])
		responseMsg = "Canceled watch on meeting ID `" + data.Values[0] + "`."
	} else {
		builder := new(strings.Builder)
		builder.WriteString("Canceled watch on the following meetings:")
		for _, meetingID := range data.Values {
			o.CancelWatch(i.GuildID, meetingID)
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
