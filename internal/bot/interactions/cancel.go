package interactions

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/angelajfisher/zoom-bot/internal/types"
)

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
