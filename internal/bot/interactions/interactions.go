package interactions

import (
	"github.com/bwmarrin/discordgo"
)

func InteractionList() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
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
				{
					Name:        "join_link",
					Description: "Link for others to join the meeting",
					Type:        discordgo.ApplicationCommandOptionString,
				},
				{
					Name:        "summary",
					Description: "Display meeting stats after it ends (default: false)",
					Type:        discordgo.ApplicationCommandOptionBoolean,
				},
				{
					Name:        "keep_history",
					Description: "How often new messages are sent / old ones deleted (default: Partial)",
					Type:        discordgo.ApplicationCommandOptionString,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Full", Value: FullHistory},
						{Name: "Partial", Value: PartialHistory},
						{Name: "Minimal", Value: MinimalHistory},
					},
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
		}, {
			Name:        "status",
			Description: "Check the status of your ongoing watch(es)",
		},
	}
}

type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

func ParseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) optionMap {
	om := make(optionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return om
}
