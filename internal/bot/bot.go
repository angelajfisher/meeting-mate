package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/angelajfisher/meeting-mate/internal/bot/interactions"
	"github.com/angelajfisher/meeting-mate/internal/types"
	"github.com/bwmarrin/discordgo"
)

var (
	BotToken string
	AppID    string
	session  *discordgo.Session
)

func Run() error {
	var err error
	session, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		return fmt.Errorf("invalid bot parameters: %w", err)
	}

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	session.AddHandler(func(_ *discordgo.Session, r *discordgo.Ready) {
		log.Println("Logged in as", r.User.String())
	})

	_, err = session.ApplicationCommandBulkOverwrite(AppID, "", interactions.InteractionList)
	if err != nil {
		return fmt.Errorf("could not register bot commands: %w", err)
	}

	err = session.Open()
	if err != nil {
		return fmt.Errorf("could not open bot session: %w", err)
	}

	if err = session.UpdateCustomStatus("Check the status of your watches with /status"); err != nil {
		log.Printf("could not set custom status: %s", err)
	}

	return nil
}

func Stop() error {
	if session == nil {
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

	err := session.Close()
	if err != nil {
		return fmt.Errorf("could not close session gracefully: %w", err)
	}

	fmt.Print("Done!\n")
	return nil
}
