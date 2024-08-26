package interactions

import (
	"log"

	"github.com/angelajfisher/zoom-bot/internal/types"
	"github.com/bwmarrin/discordgo"
)

func HandleCancel(s *discordgo.Session, i *discordgo.InteractionCreate, _ optionMap) {
	log.Printf("%s: /cancel ID %s", i.Member.User, meetingID)

	types.WatchMeetingID <- types.Canceled

	var content string
	if meetingID == "" {
		content = "Nothing to cancel: no meetings are currently being watched."
	} else {
		content = "Canceled watch on meeting " + meetingID + "."
		meetingID = ""
	}

	err := s.UpdateCustomStatus("")
	if err != nil {
		log.Printf("could not set custom status: %s", err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Printf("could not respond to interaction: %s", err)
	}
}
