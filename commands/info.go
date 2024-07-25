package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleInfo(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{{
				Type:        discordgo.EmbedTypeRich,
				Title:       "Current Weather",
				Description: "Weather for Seattle, WA",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Conditions",
						Value:  "Sunny",
						Inline: true,
					},
					{
						Name:   "Temperature",
						Value:  "74°F",
						Inline: true,
					},
					{
						Name:   "Humidity",
						Value:  "64%",
						Inline: true,
					},
					{
						Name:   "Wind",
						Value:  "8 mph",
						Inline: true,
					},
				},
			}},
		},
	})

	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}
