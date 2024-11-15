package interactions

import (
	"github.com/bwmarrin/discordgo"
)

const (
	WATCH_COMMAND   = "watch"
	CANCEL_COMMAND  = "cancel"
	STATUS_COMMAND  = "status"
	FULL_HISTORY    = "Full"    // No old meeting messages are removed
	PARTIAL_HISTORY = "Partial" // Keep the old meeting message only if it's been buried by conversation
	MINIMAL_HISTORY = "Minimal" // Do not keep any old meeting messages
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
						{Name: FULL_HISTORY, Value: FULL_HISTORY},
						{Name: PARTIAL_HISTORY, Value: PARTIAL_HISTORY},
						{Name: MINIMAL_HISTORY, Value: MINIMAL_HISTORY},
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
