package interactions

import (
	"log"
	"strings"

	"github.com/angelajfisher/zoom-bot/internal/types"
	"github.com/bwmarrin/discordgo"
)

func HandleWatch(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {

	meetingID := opts["meeting_id"].StringValue()

	silent := true
	if v, ok := opts["silent"]; ok && !v.BoolValue() {
		silent = v.BoolValue()
	}

	builder := new(strings.Builder)

	builder.WriteString("Initiating watch on meeting ID " + meetingID + "!\nStop at any time with /cancel")

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
		},
	})
	if err != nil {
		log.Printf("could not respond to interaction: %s", err)
	}

	types.WatchMeetingID <- meetingID

	var content discordgo.WebhookParams
	if silent {
		content.Flags = discordgo.MessageFlagsSuppressNotifications
	}

	err = s.UpdateCustomStatus("Watching meeting ID " + meetingID)
	if err != nil {
		log.Printf("could not set custom status: %s", err)
	}

	var (
		meetingStatusMsg *discordgo.Message
		meetingEnded     = true
		participantsList = make(map[string]string) // map[participantID]participantName
	)
	for {
		zoomData := <-types.MeetingData
		if zoomData.EventType == types.Canceled {
			break
		}

		if meetingEnded {
			content.Embeds = []*discordgo.MessageEmbed{{
				Type:        discordgo.EmbedTypeRich,
				Title:       zoomData.MeetingName,
				Description: "This meeting is ongoing.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Current Participants",
						Value: stringifyParticipants(participantsList),
					},
				},
			}}
		}

		switch zoomData.EventType {
		case types.ParticipantJoin:
			meetingEnded = false
			participantsList[zoomData.ParticipantID] = zoomData.ParticipantName
			content.Embeds[0].Fields[0].Value = stringifyParticipants(participantsList)

		case types.ParticipantLeave:
			delete(participantsList, zoomData.ParticipantID)
			content.Embeds[0].Fields[0].Value = stringifyParticipants(participantsList)

		case types.MeetingEnd:
			clear(participantsList)
			content.Embeds[0].Description = "This meeting has ended."
			content.Embeds[0].Fields = nil
			meetingEnded = true
		}

		if meetingStatusMsg != nil {
			updatedContent := discordgo.WebhookEdit{Embeds: &content.Embeds}
			meetingStatusMsg, err = s.FollowupMessageEdit(i.Interaction, meetingStatusMsg.ID, &updatedContent)
			if err != nil {
				log.Printf("could not respond to interaction: %s", err)
			}
			if meetingEnded {
				meetingStatusMsg = nil
			}
		} else {
			meetingStatusMsg, err = s.FollowupMessageCreate(i.Interaction, true, &content)
			if err != nil {
				log.Printf("could not respond to interaction: %s", err)
			}
		}
	}

	if meetingStatusMsg != nil {
		content.Embeds[0].Description = "**Status Unknown**\nThe watch on this meeting was canceled."
		content.Embeds[0].Fields = nil

		updatedContent := discordgo.WebhookEdit{Embeds: &content.Embeds}
		_, err = s.FollowupMessageEdit(i.Interaction, meetingStatusMsg.ID, &updatedContent)
		if err != nil {
			log.Printf("could not respond to interaction: %s", err)
		}
	}

}

func stringifyParticipants(participants map[string]string) string {
	participantStr := ""
	for _, name := range participants {
		participantStr += name + "\n"
	}
	return participantStr
}

func HandleCancel(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {
	types.WatchMeetingID <- types.Canceled

	err := s.UpdateCustomStatus("")
	if err != nil {
		log.Printf("could not set custom status: %s", err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Canceled meeting watch.",
		},
	})
	if err != nil {
		log.Printf("could not respond to interaction: %s", err)
	}
}
