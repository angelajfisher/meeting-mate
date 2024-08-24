package interactions

import (
	"log"
	"strings"

	"github.com/angelajfisher/zoom-bot/internal/types"
	"github.com/bwmarrin/discordgo"
)

var (
	Watchers  []chan struct{} // Shutdown communication channels for all active meeting watches
	meetingID string
)

func HandleWatch(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {
	var (
		err               error
		newMeetingID      = opts["meeting_id"].StringValue()
		sendSilently      = true
		responseMsg       string
		shutdownCh        = make(chan struct{})
		meetingStatusRes  *discordgo.Message
		meetingInProgress = false
		participantsList  = make(map[string]string) // map[participantID]participantName
	)
	log.Printf("[%s] %s: /watch ID %s", types.CurrentTime(), i.Member.User, newMeetingID)

	if v, ok := opts["silent"]; ok && !v.BoolValue() {
		sendSilently = v.BoolValue()
	}

	if newMeetingID == meetingID {
		responseMsg = "Watch on meeting ID " + newMeetingID + " is already ongoing."
		// TODO: make this response ephemeral, then disregard request
	} else {
		meetingID = newMeetingID
		responseMsg = "Initiating watch on meeting ID " + newMeetingID + "!\nStop at any time with /cancel"
	}

	Watchers = append(Watchers, shutdownCh)
	go func() {
		<-shutdownCh
		types.WatchMeetingID <- types.Shutdown
	}()

	if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: responseMsg,
		},
	}); err != nil {
		log.Printf("could not respond to interaction: %s", err)
	}
	types.WatchMeetingID <- meetingID

	if err = s.UpdateCustomStatus("Watching meeting ID " + meetingID); err != nil {
		log.Printf("could not set custom status: %s", err)
	}

	meetingMsgContent := discordgo.WebhookParams{Embeds: []*discordgo.MessageEmbed{{Type: discordgo.EmbedTypeRich}}}
	if sendSilently {
		meetingMsgContent.Flags = discordgo.MessageFlagsSuppressNotifications
	}

	for zoomData := range types.MeetingData {
		if zoomData.EventType == types.Canceled {
			meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch on this meeting was canceled."
			break
		} else if zoomData.EventType == types.Shutdown {
			meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch has stopped due to bot shutdown."
			break
		}

		// If there wasn't a meeting in progress before this data came in, start a new meeting message
		if !meetingInProgress {
			meetingInProgress = true
			meetingMsgContent.Embeds[0].Title = zoomData.MeetingName
			meetingMsgContent.Embeds[0].Description = "This meeting is ongoing."
			meetingMsgContent.Embeds[0].Fields = []*discordgo.MessageEmbedField{{Name: "Current Participants"}}
		}

		switch zoomData.EventType {
		case types.ParticipantJoin:
			participantsList[zoomData.ParticipantID] = zoomData.ParticipantName
			meetingMsgContent.Embeds[0].Fields[0].Value = stringifyParticipants(participantsList)

		case types.ParticipantLeave:
			delete(participantsList, zoomData.ParticipantID)
			meetingMsgContent.Embeds[0].Fields[0].Value = stringifyParticipants(participantsList)

		case types.MeetingEnd:
			clear(participantsList)
			meetingMsgContent.Embeds[0].Description = "This meeting has ended."
			meetingMsgContent.Embeds[0].Fields = nil
			meetingInProgress = false
		}

		if meetingStatusRes != nil {
			updatedContent := discordgo.MessageEdit{
				Embeds:  &meetingMsgContent.Embeds,
				ID:      meetingStatusRes.ID,
				Channel: meetingStatusRes.ChannelID,
			}
			if meetingStatusRes, err = s.ChannelMessageEditComplex(&updatedContent); err != nil {
				log.Printf("could not respond to interaction: %s", err)
			}
			if !meetingInProgress {
				meetingStatusRes = nil
			}
		} else {
			if meetingStatusRes, err = s.FollowupMessageCreate(i.Interaction, true, &meetingMsgContent); err != nil {
				log.Printf("could not respond to interaction: %s", err)
			}
		}
	}

	if meetingStatusRes != nil {
		meetingMsgContent.Embeds[0].Fields = nil
		updatedContent := discordgo.MessageEdit{
			Embeds:  &meetingMsgContent.Embeds,
			ID:      meetingStatusRes.ID,
			Channel: meetingStatusRes.ChannelID,
		}
		if _, err = s.ChannelMessageEditComplex(&updatedContent); err != nil {
			log.Printf("could not respond to interaction: %s", err)
		}
	}
}

func stringifyParticipants(participants map[string]string) string {
	builder := new(strings.Builder)
	for _, name := range participants {
		builder.WriteString(name + "\n")
	}
	if builder.String() == "" {
		builder.WriteString("This meeting appears to be empty")
	}
	return builder.String()
}
