package interactions

import (
	"github.com/bwmarrin/discordgo"
)

var List = []*discordgo.ApplicationCommand{
	{
		Name:        "info",
		Description: "Get the bot to display some info via embed",
	}, {
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
				Description: "Post status updates silently (no ping) - default: true",
				Type:        discordgo.ApplicationCommandOptionBoolean,
			},
		},
	}, {
		Name:        "cancel",
		Description: "Cancel the current meeting watch",
	},
}

type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

func ParseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (om optionMap) {
	om = make(optionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return
}
