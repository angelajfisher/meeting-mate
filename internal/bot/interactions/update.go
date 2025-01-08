package interactions

import (
	"log"
	"net/url"

	"github.com/angelajfisher/meeting-mate/internal/orchestrator"
	"github.com/bwmarrin/discordgo"
)

func HandleUpdate(s *discordgo.Session, i *discordgo.InteractionCreate, o orchestrator.Orchestrator, opts optionMap) {
	meetingID := opts[MEETING_OPT].StringValue()
	log.Printf("%s: /update ID %s in %s", i.Member.User, meetingID, i.GuildID)

	var invalidResponseMsg string

	// Verify that the requested meeting ID exists
	if !o.IsOngoingWatch(i.GuildID, meetingID) {
		invalidResponseMsg = "Nothing to update: meeting ID `" + meetingID + "` isn't being watched in this server."
	}

	// Check for valid join link if provided
	if v, ok := opts[LINK_OPT]; ok && invalidResponseMsg == "" {
		u, parseErr := url.Parse(v.StringValue())
		if parseErr != nil || u.Scheme != "https" {
			invalidResponseMsg = "Invalid join link provided. Please ensure your URL is correct and starts with \"https://\""
		}
	}

	if invalidResponseMsg != "" {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: invalidResponseMsg,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("HandleUpdate-Invalid: could not respond to interaction: %s", err)
		}
		return
	}

	newFlags := generateWatchFlags(opts)
	o.UpdateFlags(i.GuildID, meetingID, newFlags)

	silent := "True"
	if !newFlags.Silent {
		silent = "False"
	}
	summaries := "True"
	if !newFlags.Summaries {
		summaries = "False"
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Successfully updated! The watch on meeting ID `" + meetingID +
				"` now has the following options selected:\n\n**Silent**: `" +
				silent + "`\n**Join link**: " + func() string {
				if newFlags.JoinLink == "" {
					return "n/a"
				}
				return "`" + newFlags.JoinLink + "`"
			}() + "\n**Summaries**: `" + summaries + "`\n**History level**: `" + newFlags.HistoryLevel + "`",
		},
	})
	if err != nil {
		log.Printf("HandleUpdate-Success: could not respond to interaction: %s", err)
	}
}
