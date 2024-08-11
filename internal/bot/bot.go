package bot

import (
	"log"
	"os"
	"os/signal"

	"github.com/angelajfisher/zoom-bot/internal/interactions"
	"github.com/bwmarrin/discordgo"
)

var (
	BotToken string
	AppID    string
)

func Run() {
	session, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("ERROR: Invalid bot parameters: %v", err)
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
		log.Fatalf("ERROR: Could not register bot commands: %s", err)
	}

	err = session.Open()
	if err != nil {
		log.Fatalf("ERROR: Could not open bot session: %s", err)
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	err = session.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}
}
