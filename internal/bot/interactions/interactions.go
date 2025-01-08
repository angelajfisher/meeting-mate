package interactions

import (
	"strings"

	"github.com/angelajfisher/meeting-mate/internal/types"
	"github.com/bwmarrin/discordgo"
)

const (
	// Command IDs
	WATCH_COMMAND  = "watch"
	CANCEL_COMMAND = "cancel"
	STATUS_COMMAND = "status"
	UPDATE_COMMAND = "update"

	// Watch option flags
	MEETING_OPT = "meeting_id"
	SILENT_OPT  = "silent"
	LINK_OPT    = "join_link"
	SUMMARY_OPT = "summary"
	HISTORY_OPT = "keep_history"
)

func InteractionList() []*discordgo.ApplicationCommand {
	watchOptions := watchOptions()
	return []*discordgo.ApplicationCommand{
		{
			Name:        WATCH_COMMAND,
			Description: "Begin watching a meeting's participant list",
			Options:     watchOptions,
		}, {
			Name:        CANCEL_COMMAND,
			Description: "Cancel the watch on a meeting",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        MEETING_OPT,
					Description: "ID of the Zoom meeting",
					Type:        discordgo.ApplicationCommandOptionString,
				},
			},
		}, {
			Name:        STATUS_COMMAND,
			Description: "Check the status of your ongoing watch(es)",
		}, {
			Name:        UPDATE_COMMAND,
			Description: "Update the options on an ongoing watch",
			Options:     watchOptions,
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

func watchOptions() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Name:        MEETING_OPT,
			Description: "ID of the Zoom meeting",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
		{
			Name:        SILENT_OPT,
			Description: "Post status updates @silent-ly (default: true)",
			Type:        discordgo.ApplicationCommandOptionBoolean,
		},
		{
			Name:        LINK_OPT,
			Description: "Link for others to join the meeting",
			Type:        discordgo.ApplicationCommandOptionString,
		},
		{
			Name:        SUMMARY_OPT,
			Description: "Display meeting stats after it ends (default: true)",
			Type:        discordgo.ApplicationCommandOptionBoolean,
		},
		{
			Name:        HISTORY_OPT,
			Description: "How often new messages are sent / old ones deleted (default: Partial)",
			Type:        discordgo.ApplicationCommandOptionString,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: types.FULL_HISTORY, Value: types.FULL_HISTORY},
				{Name: types.PARTIAL_HISTORY, Value: types.PARTIAL_HISTORY},
				{Name: types.MINIMAL_HISTORY, Value: types.MINIMAL_HISTORY},
			},
		},
	}
}

func generateWatchFlags(opts optionMap) types.FeatureFlags {
	// Begin building new restart command
	builder := new(strings.Builder)
	builder.WriteString("```/watch meeting_id: " + opts[MEETING_OPT].StringValue())

	// Store flag choices
	flags := types.FeatureFlags{
		Silent: func() bool {
			if v, exists := opts[SILENT_OPT]; exists && !v.BoolValue() {
				builder.WriteString(" " + SILENT_OPT + ": false")
				return false
			}
			return true
		}(),
		JoinLink: func() string {
			if v, exists := opts[LINK_OPT]; exists && v.StringValue() != "" {
				builder.WriteString(" " + LINK_OPT + ": " + v.StringValue())
				return v.StringValue()
			}
			return ""
		}(),
		Summaries: func() bool {
			if v, exists := opts[SUMMARY_OPT]; exists && !v.BoolValue() {
				builder.WriteString(" " + SUMMARY_OPT + ": false")
				return false
			}
			return true
		}(),
		HistoryLevel: func() string {
			if v, exists := opts[HISTORY_OPT]; exists && v.StringValue() != types.PARTIAL_HISTORY {
				builder.WriteString(" " + HISTORY_OPT + ": " + v.StringValue())
				return v.StringValue()
			}
			return types.PARTIAL_HISTORY
		}(),
		RestartCommand: func() string {
			builder.WriteString("```")
			return builder.String()
		}(),
	}

	return flags
}
