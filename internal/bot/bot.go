package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/angelajfisher/meeting-mate/internal/bot/interactions"
	"github.com/angelajfisher/meeting-mate/internal/types"
	"github.com/bwmarrin/discordgo"
)

type Config struct {
	BotToken string
	AppID    string
	session  *discordgo.Session
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
				interactions.HandleCancelSelection(s, i)
			}
			return
		}

		data := i.ApplicationCommandData()
		switch data.Name {
		case "watch":
			interactions.HandleWatch(s, i, interactions.ParseOptions(data.Options))
		case "cancel":
			interactions.HandleCancel(s, i, interactions.ParseOptions(data.Options))
		case "status":
			interactions.HandleStatus(s, i)
		}
	})

	bc.session.AddHandler(func(_ *discordgo.Session, r *discordgo.Ready) {
		log.Println("Logged in as", r.User.String())
	})

	interacts := interactions.InteractionList()
	_, err = bc.session.ApplicationCommandBulkOverwrite(bc.AppID, "", interacts)
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

	return nil
}

func Stop(bc *Config) error {
	if bc.session == nil {
		return nil
	}

	fmt.Print("Bot shutting down...")

	// Notify all active watchers of shutdown
	for _, meeting := range types.AllMeetings.Meetings {
		for _, watch := range types.DataListeners.GetMeetingListeners(meeting.GetID()) {
			watch <- types.EventData{EventType: types.BotShutdown}
		}
	}

	// Give watchers time to stop
	time.Sleep(time.Second)

	err := bc.session.Close()
	if err != nil {
		return fmt.Errorf("could not close session gracefully: %w", err)
	}

	fmt.Print("Done!\n")
	return nil
}
