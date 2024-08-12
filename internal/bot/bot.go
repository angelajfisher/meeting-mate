package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"

	"github.com/angelajfisher/zoom-bot/internal/interactions"
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
		return fmt.Errorf("Invalid bot parameters: %v", err)
	}

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		data := i.ApplicationCommandData()
		switch data.Name {
		case "info":
			interactions.HandleInfo(s, i, interactions.ParseOptions(data.Options))
		case "watch":
			interactions.HandleWatch(s, i, interactions.ParseOptions(data.Options))
		case "cancel":
			interactions.HandleCancel(s, i, interactions.ParseOptions(data.Options))
		}
	})

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Logged in as", r.User.String())
	})

	_, err = session.ApplicationCommandBulkOverwrite(AppID, "", interactions.List)
	if err != nil {
		return fmt.Errorf("Could not register bot commands: %v", err)
	}

	err = session.Open()
	if err != nil {
		return fmt.Errorf("Could not open bot session: %v", err)
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch
	return nil

}

func Stop() error {

	if session == nil {
		return nil
	}

	fmt.Print("Bot shutting down...")

	// Notify all active watchers of shutdown
	for _, c := range interactions.Watchers {
		c <- false
	}

	err := session.Close()
	if err != nil {
		return fmt.Errorf("Could not close bot session gracefully: %v", err)
	}

	fmt.Print("Done!\n")
	return nil

}
