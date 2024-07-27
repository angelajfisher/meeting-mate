package main

import (
	"log"
	"os"
	"sync"

	"github.com/angelajfisher/zoom-bot/bot"
	"github.com/angelajfisher/zoom-bot/webhooks"
)

func main() {
	webhooks.Secret = os.Getenv("ZOOM_TOKEN")
	bot.BotToken = os.Getenv("BOT_TOKEN")
	bot.AppID = os.Getenv("APP_ID")

	if bot.BotToken == "" || bot.AppID == "" || webhooks.Secret == "" {
		log.Fatal("ERROR: Please ensure your ZOOM_TOKEN, BOT_TOKEN, and APP_ID variables are in the environment before building.")
	}

	// TODO: Configure channels to share incoming info from Zoom w/ bot
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		webhooks.Listen()
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		bot.Run()
		wg.Done()
	}()
	wg.Wait()

}
