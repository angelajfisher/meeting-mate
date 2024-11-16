package interactions

import (
	"github.com/angelajfisher/meeting-mate/internal/types"
	"github.com/bwmarrin/discordgo"
)

const (
	WATCH_COMMAND   = "watch"
	CANCEL_COMMAND  = "cancel"
	STATUS_COMMAND  = "status"
)

func InteractionList() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        WATCH_COMMAND,
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
					Description: "Display meeting stats after it ends (default: true)",
					Type:        discordgo.ApplicationCommandOptionBoolean,
				},
				{
					Name:        "keep_history",
					Description: "How often new messages are sent / old ones deleted (default: Partial)",
					Type:        discordgo.ApplicationCommandOptionString,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: types.FULL_HISTORY, Value: types.FULL_HISTORY},
						{Name: types.PARTIAL_HISTORY, Value: types.PARTIAL_HISTORY},
						{Name: types.MINIMAL_HISTORY, Value: types.MINIMAL_HISTORY},
					},
				},
			},
		}, {
			Name:        CANCEL_COMMAND,
			Description: "Cancel the watch on a meeting",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "meeting_id",
					Description: "ID of the Zoom meeting",
					Type:        discordgo.ApplicationCommandOptionString,
				},
			},
		}, {
			Name:        STATUS_COMMAND,
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
