package interactions

import (
	"log"
	"strings"

	"github.com/angelajfisher/meeting-mate/internal/utils"
	"github.com/bwmarrin/discordgo"
)

func HandleStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var (
		activeWatches = utils.MeetingWatches.GetMeetings(i.GuildID)
		response      string
	)

	if len(activeWatches) == 0 {
		response = "There are no ongoing watches in this server. Get one started with `/watch`!"
	} else {
		meetingIDs := make([]string, len(activeWatches))
		i := 0
		for id := range activeWatches {
			meetingIDs[i] = id
			i++
		}
		if len(activeWatches) == 1 {
			response = "There is an ongoing watch on meeting ID `" + meetingIDs[0] + "`."
		} else {
			builder := new(strings.Builder)
			builder.WriteString("The following meeting IDs have ongoing watches:")
			for _, id := range meetingIDs {
				builder.WriteString("\n- `" + id + "`")
			}
			response = builder.String()
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("could not respond to interaction: %s", err)
	}
}
