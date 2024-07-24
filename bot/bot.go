package bot

import (
	"log"
	"os"
	"os/signal"

	"github.com/angelajfisher/zoom-bot/commands"
	"github.com/bwmarrin/discordgo"
)

var (
	BotToken string
	AppID    string
)

func Run() {
	session, _ := discordgo.New("Bot " + BotToken)

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		data := i.ApplicationCommandData()
		if data.Name != "echo" {
			return
		}

		commands.HandleEcho(s, i, commands.ParseOptions(data.Options))
	})

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})

	_, err := session.ApplicationCommandBulkOverwrite(AppID, "", commands.List)
	if err != nil {
		log.Fatalf("could not register commands: %s", err)
	}

	err = session.Open()
	if err != nil {
		log.Fatalf("could not open session: %s", err)
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	err = session.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}
}
