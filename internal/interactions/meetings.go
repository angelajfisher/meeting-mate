package interactions

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/angelajfisher/zoom-bot/internal/data"
	"github.com/bwmarrin/discordgo"
)

func HandleWatch(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {

	silent := true
	if v, ok := opts["silent"]; ok && !v.BoolValue() {
		silent = v.BoolValue()
	}

	builder := new(strings.Builder)

	builder.WriteString("Initiating watch on meeting ID " + opts["meeting_id"].StringValue() + "! Stop at any time with /cancel")

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
		},
	})
	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}

	data.WatchMeetingID <- opts["meeting_id"].StringValue()

	var content discordgo.WebhookParams
	if silent {
		content.Flags = discordgo.MessageFlagsSuppressNotifications
	}
	log.Println("Listening to data channel!")
	for {
		zoomData := <-data.MeetingData

		log.Printf("Data received from channel: %v\n", zoomData)

		data, err := json.Marshal(zoomData)
		if err != nil {
			log.Panicf("could not marshal zoomData: %s", err)
		}
		content.Content = string(data)

		_, err = s.FollowupMessageCreate(i.Interaction, true, &content)
		if err != nil {
			log.Panicf("could not respond to interaction: %s", err)
		}
	}

}
