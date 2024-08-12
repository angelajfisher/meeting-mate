package bot

import (
	"fmt"
	"log"

	"github.com/angelajfisher/zoom-bot/internal/bot/interactions"
	"github.com/bwmarrin/discordgo"
)

var (
	BotToken string
	AppID    string
	session  *discordgo.Session
	stop     chan struct{}
)

func Run() error {
	var err error
	session, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		return fmt.Errorf("invalid bot parameters: %w", err)
	}

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		data := i.ApplicationCommandData()
		switch data.Name {
		case "watch":
			interactions.HandleWatch(s, i, interactions.ParseOptions(data.Options))
		case "cancel":
			interactions.HandleCancel(s, i, interactions.ParseOptions(data.Options))
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

	// Listen for shutdown signal
	<-stop
	err = session.Close()
	if err != nil {
		return fmt.Errorf("could not close bot session gracefully: %w", err)
	}
	return nil
}

func Stop() error {
	if session == nil {
		return nil
	}

	fmt.Print("Bot shutting down...")

	// Notify all active watchers of shutdown
	for _, c := range interactions.Watchers {
		c <- struct{}{}
	}

	go func() {
		stop <- struct{}{}
	}()

	fmt.Print("Done!\n")
	return nil
}
