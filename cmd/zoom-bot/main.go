package main

import (
	"log"
	"os"

	"github.com/angelajfisher/zoom-bot/bot"
)

func main() {
	bot.BotToken = os.Getenv("BOT_TOKEN")
	bot.AppID = os.Getenv("APP_ID")

	if bot.BotToken == "" || bot.AppID == "" {
		log.Fatal("Please ensure your BotToken and AppID variables are in the environment before building.")
	}

	bot.Run()
}
