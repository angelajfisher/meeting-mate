package interactions

import (
	"github.com/bwmarrin/discordgo"
)

var InteractionList = []*discordgo.ApplicationCommand{
	{
		Name:        "watch",
		Description: "Begin watching a meeting's participant list",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "meeting_id",
				Description: "ID of the Zoom meeting",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "silent",
				Description: "Post status updates @silent-ly (default: true)",
				Type:        discordgo.ApplicationCommandOptionBoolean,
			},
		},
	}, {
		Name:        "cancel",
		Description: "Cancel the watch on a meeting",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "meeting_id",
				Description: "ID of the Zoom meeting",
				Type:        discordgo.ApplicationCommandOptionString,
			},
		},
	},
}

type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

func ParseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) optionMap {
	om := make(optionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return om
}
