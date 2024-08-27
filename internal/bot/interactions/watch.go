package interactions

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/angelajfisher/meeting-mate/internal/types"
	"github.com/bwmarrin/discordgo"
)

// TODO: All this should be within a struct; this is a mess

var (
	Watchers         []chan struct{}           // Shutdown communication channels for all active meeting watches
	participantsList = make(map[string]string) // map[participantID]participantName)
	meetingID        string
)

func HandleWatch(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {
	var (
		err          error
		newMeetingID = opts["meeting_id"].StringValue()
		sendSilently = true
		responseMsg  string
		shutdownCh   = make(chan struct{})
	)
	log.Printf("%s: /watch ID %s", i.Member.User, newMeetingID)

	if v, ok := opts["silent"]; ok && !v.BoolValue() {
		sendSilently = v.BoolValue()
	}

	switch {
	case newMeetingID == "":
		responseMsg = "Please supply a meeting ID to watch."

	case newMeetingID == meetingID:
		if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Watch on meeting ID " + newMeetingID + " is already ongoing.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			log.Printf("could not respond to interaction: %s", err)
		}
		return

	case meetingID != "":
		meetingID = newMeetingID
		responseMsg = "Canceling watch on meeting ID " + meetingID +
			" and starting new watch on meeting ID " + newMeetingID + "!\nStop at any time with /cancel"

	case meetingID == "":
		meetingID = newMeetingID
		responseMsg = "Initiating watch on meeting ID " + newMeetingID + "!\nStop at any time with /cancel"
	}

	Watchers = append(Watchers, shutdownCh)
	go func() {
		<-shutdownCh
		if err = s.UpdateCustomStatus(""); err != nil {
			log.Printf("could not set custom status: %s", err)
		}
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

	meetingMsgContent := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{{Type: discordgo.EmbedTypeRich,
		Description: "Loading..."}}}
	if sendSilently {
		meetingMsgContent.Flags = discordgo.MessageFlagsSuppressNotifications
	}

	var meetingStatusRes *discordgo.Message

	for zoomData := range types.MeetingData {
		if zoomData.EventType == types.Canceled {
			meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch on this meeting was canceled."
			break
		} else if zoomData.EventType == types.Shutdown {
			meetingMsgContent.Embeds[0].Description = "**Status Unknown**\nThe watch stopped due to bot shutdown." +
				" Please restart with /watch when available."
			break
		}

		if meetingStatusRes == nil {
			meetingStatusRes, err = s.ChannelMessageSendComplex(i.Interaction.ChannelID, meetingMsgContent)
			if err != nil {
				log.Printf("could not respond to interaction: %s", err)
			}
		}

		meetingStatusRes, err = updateMeetingMsg(s, zoomData, meetingMsgContent, meetingStatusRes)
		if err != nil {
			log.Println(err)
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

func updateMeetingMsg(
	s *discordgo.Session,
	zoomData types.EventData,
	meetingMsgContent *discordgo.MessageSend,
	meetingStatusRes *discordgo.Message,
) (*discordgo.Message, error) {
	var (
		meetingInProgress = false
		err               error
	)

	// If there wasn't a meeting in progress before this data came in, start a new meeting message
	if !meetingInProgress {
		meetingInProgress = true
		meetingMsgContent.Embeds[0].Title = zoomData.MeetingName
		meetingMsgContent.Embeds[0].Description = "This meeting is in progress."
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
		meetingMsgContent.Embeds[0].Description = "This meeting ended."
		meetingMsgContent.Embeds[0].Fields = nil
		meetingInProgress = false
	}

	if meetingStatusRes != nil {
		updatedContent := discordgo.MessageEdit{
			Embeds:  &meetingMsgContent.Embeds,
			ID:      meetingStatusRes.ID,
			Channel: meetingStatusRes.ChannelID,
		}
		meetingStatusRes, err = s.ChannelMessageEditComplex(&updatedContent)
		if err != nil {
			return meetingStatusRes, fmt.Errorf("could not respond to interaction: %w", err)
		}
		if !meetingInProgress {
			meetingStatusRes = nil
		}
	} else {
		return meetingStatusRes, errors.New("could not update meeting status: meeting message is nil")
	}

	return meetingStatusRes, nil
}

func stringifyParticipants(participants map[string]string) string {
	builder := new(strings.Builder)
	for _, name := range participants {
		builder.WriteString(name + "\n")
	}
	if builder.String() == "" {
		builder.WriteString("Participants unknown")
	}
	return builder.String()
}
