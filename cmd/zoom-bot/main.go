package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/angelajfisher/zoom-bot/internal/bot"
	"github.com/angelajfisher/zoom-bot/internal/server"
	"github.com/joho/godotenv"
)

func main() {

	devMode := flag.Bool("dev", false, "run the program in development mode")
	envPath := flag.String("envFile", "", "program will load environment variables from the file at this path")
	staticDir := flag.String("staticDir", "./static", "path to static directory containing site files")
	webhookPort := flag.String("webhookPort", ":12345", "port at which the webhook server will listen for incoming hooks - default :12345")
	flag.Parse()

	if *devMode {
		log.Println("WARN: Initializing Zoom Boot in DEVELOPMENT mode")
	}

	if *envPath != "" {
		err := godotenv.Load(*envPath)
		if err != nil {
			log.Fatal("ERROR: Could not load .env file at provided path.")
		}
	}

	server.BaseURL = "/projects/zoom-bot"
	server.StaticDir = *staticDir
	server.Port = *webhookPort
	server.Secret = os.Getenv("ZOOM_TOKEN")
	bot.BotToken = os.Getenv("BOT_TOKEN")
	bot.AppID = os.Getenv("APP_ID")

	if bot.BotToken == "" || bot.AppID == "" || server.Secret == "" {
		log.Fatal("ERROR: Please ensure your ZOOM_TOKEN, BOT_TOKEN, and APP_ID variables are in the environment before building.")
	}

	// TODO: Configure channels to share incoming info from Zoom w/ bot
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		server.Start(*devMode)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		bot.Run()
		wg.Done()
	}()
	wg.Wait()

}
