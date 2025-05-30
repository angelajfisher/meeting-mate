package interactions

import (
	"log"
	"strings"

	"github.com/angelajfisher/meeting-mate/internal/orchestrator"
	"github.com/bwmarrin/discordgo"
)

const (
	CANCEL_ID = "meeting_cancel_selection_"
	noWatch   = "Nothing to cancel: there is no ongoing watch on this meeting."
)

// Handles the initial `/cancel` command
func HandleCancel(s *discordgo.Session, i *discordgo.InteractionCreate, o orchestrator.Orchestrator, opts optionMap) {
	var (
		err       error
		meetingID string
	)

	// Check for a meeting ID provided with the command
	if v, ok := opts[MEETING_OPT]; ok && v.StringValue() != "" {
		meetingID = v.StringValue()
		log.Printf("%s in %s: /cancel ID %s", i.Member.User, i.GuildID, meetingID)
	} else {
		log.Printf("%s in %s: /cancel", i.Member.User, i.GuildID)
	}

	//
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
			log.Printf("HandleCancel-IDProvided: could not respond to interaction: %s", err)
		}
		return
	}

	//
	// ID not provided path

	// Format all ongoing meeting watches in this server into selectable options
	guildWatches := []discordgo.SelectMenuOption{}
	for _, meeting := range o.GetGuildMeetings(i.GuildID) {
		meetingLabel := meeting
		meetingName := o.GetMeetingName(meeting)
		if meetingName != "" {
			meetingLabel += " (" + meetingName + ")"
		}
		guildWatches = append(guildWatches, discordgo.SelectMenuOption{
			Label: meetingLabel,
			Value: meeting,
		})
	}

	if len(guildWatches) == 0 {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Nothing to cancel: there are no ongoing watches in this server.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("HandleCancel-NoWatches: could not respond to interaction: %s", err)
		}
		return
	}

	minVals := 1
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Which ongoing meeting watches would you like to cancel?",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						MenuType:    discordgo.StringSelectMenu,
						Options:     guildWatches,
						MinValues:   &minVals,
						MaxValues:   len(guildWatches),
						CustomID:    CANCEL_ID + i.Interaction.Member.User.ID,
						Placeholder: "Select meeting ID(s)",
					},
				}},
			},
			CustomID: "meeting_cancel_modal_" + i.Interaction.Member.User.ID,
		},
	})
	if err != nil {
		log.Printf("HandleCancel-IDNotProvided: could not respond to interaction: %s", err)
	}
}

// Handles the user response to the multiselect menu returned by /cancel when an ID is not provided
func HandleCancelSelection(s *discordgo.Session, i *discordgo.InteractionCreate, o orchestrator.Orchestrator) {
	data := i.MessageComponentData()

	var responseMsg string
	if len(data.Values) == 1 {
		o.CancelWatch(i.GuildID, data.Values[0])
		responseMsg = "Canceled watch on meeting ID `" + data.Values[0] + "`."
	} else {
		builder := new(strings.Builder)
		builder.WriteString("Canceled watches on the following meetings:")
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
		log.Printf("HandleCancelSelections: could not respond to interaction: %s", err)
	}
}
