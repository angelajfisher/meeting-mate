package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/angelajfisher/meeting-mate/internal/bot/interactions"
	"github.com/angelajfisher/meeting-mate/internal/orchestrator"
	"github.com/bwmarrin/discordgo"
)

type Config struct {
	BotToken     string
	AppID        string
	Orchestrator orchestrator.Orchestrator
	session      *discordgo.Session
}

func Run(bc *Config) error {
	var err error
	bc.session, err = discordgo.New("Bot " + bc.BotToken)
	if err != nil {
		return fmt.Errorf("invalid bot parameters: %w", err)
	}

	bc.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			if i.Type == discordgo.InteractionMessageComponent {
				interactions.HandleCancelSelection(s, i, bc.Orchestrator)
			}
			return
		}

		data := i.ApplicationCommandData()
		switch data.Name {
		case interactions.WATCH_COMMAND:
			interactions.HandleWatch(s, i, bc.Orchestrator, interactions.ParseOptions(data.Options))
		case interactions.CANCEL_COMMAND:
			interactions.HandleCancel(s, i, bc.Orchestrator, interactions.ParseOptions(data.Options))
		case interactions.STATUS_COMMAND:
			interactions.HandleStatus(s, i, bc.Orchestrator)
		case interactions.UPDATE_COMMAND:
			interactions.HandleUpdate(s, i, bc.Orchestrator, interactions.ParseOptions(data.Options))
		}
	})

	bc.session.AddHandler(func(_ *discordgo.Session, r *discordgo.Ready) {
		log.Println("Logged in as", r.User.String())
	})

	_, err = bc.session.ApplicationCommandBulkOverwrite(bc.AppID, "", interactions.InteractionList())
	if err != nil {
		return fmt.Errorf("could not register bot commands: %w", err)
	}

	err = bc.session.Open()
	if err != nil {
		return fmt.Errorf("could not open bot session: %w", err)
	}

	if err = bc.session.UpdateCustomStatus("Check the status of your watches with /status"); err != nil {
		log.Printf("could not set custom status: %s", err)
	}

	//
	// Restart previously ongoing watches from last run

	loadedWatches := bc.Orchestrator.Database.GetAllWatches()

	// Keep track of the watches happening in each channel so a restart announcement can be sent
	channelWatches := make(map[string][]string) // channelID: []meetingIDs

	// Start the watch process for each watch loaded from the database
	for _, watch := range loadedWatches {
		go interactions.LoadSavedWatch(bc.session, bc.Orchestrator, watch)
		channelWatches[watch.ChannelID] = append(channelWatches[watch.ChannelID], watch.MeetingID)
	}
	log.Println("Loaded", len(loadedWatches), "watches from database")

	// Notify every channel of the restarted watches
	for channelID, meetingIDs := range channelWatches {
		notifyOfRestart(bc.session, meetingIDs, channelID)
	}

	return nil
}

func Stop(bc *Config) error {
	if bc.session == nil {
		return nil
	}

	fmt.Print("Bot shutting down...")

	// Notify all active watchers of shutdown
	bc.Orchestrator.Shutdown()

	// Give watches time to stop
	time.Sleep(time.Second)

	err := bc.session.Close()
	if err != nil {
		return fmt.Errorf("could not close session gracefully: %w", err)
	}

	fmt.Print("Done!\n")
	return nil
}

// Sends a message to the given channel notifying of the program's (& their watches') restart
func notifyOfRestart(s *discordgo.Session, meetingIDs []string, channelID string) {
	if len(meetingIDs) == 0 {
		return
	}

	meetingList := new(strings.Builder)
	meetingList.WriteString(
		"Data for in-progress meetings has been lost, but future updates should come through as usual.\n\n",
	)

	if len(meetingIDs) == 1 {
		meetingList.WriteString("This channel's watch on meeting ID `" + meetingIDs[0] + "` has been automatically resumed.")
	} else {
		meetingList.WriteString("This channel's watches on the following meeting IDs have been automatically resumed:")
		for _, meetingID := range meetingIDs {
			meetingList.WriteString("\n- `" + meetingID + "`")
		}
	}

	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Title:       "Meeting Mate Restarted!",
			Description: meetingList.String(),
		}},
		Flags: discordgo.MessageFlagsSuppressNotifications,
	})
	if err != nil {
		log.Printf("could not send restart message to channel ID %s: %s", channelID, err)
	}
}
