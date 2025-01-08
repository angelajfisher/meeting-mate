package interactions

import (
	"log"
	"strings"

	"github.com/angelajfisher/meeting-mate/internal/orchestrator"
	"github.com/bwmarrin/discordgo"
)

func HandleStatus(s *discordgo.Session, i *discordgo.InteractionCreate, o orchestrator.Orchestrator) {
	var (
		activeWatches = o.GetGuildMeetings(i.GuildID)
		response      string
	)

	if len(activeWatches) == 0 {
		response = "There are no ongoing watches in this server. Get one started with `/watch`!"
	} else {
		builder := new(strings.Builder)
		meetingIDs := make([]string, len(activeWatches))
		i := 0
		for _, id := range activeWatches {
			meetingIDs[i] = id
			i++
		}
		if len(activeWatches) == 1 {
			builder.WriteString("There is an ongoing watch on meeting ID `" + meetingIDs[0])
			meetingName := o.GetMeetingName(meetingIDs[0])
			if meetingName != "" {
				builder.WriteString("` (" + meetingName + ").")
			} else {
				builder.WriteString("`.")
			}
		} else {
			builder.WriteString("The following meeting IDs have ongoing watches:")
			for _, id := range meetingIDs {
				builder.WriteString("\n- `" + id + "`")
				meetingName := o.GetMeetingName(id)
				if meetingName != "" {
					builder.WriteString(" (" + meetingName + ")")
				}
			}
		}
		response = builder.String()
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("HandleStatus: could not respond to interaction: %s", err)
	}
}
