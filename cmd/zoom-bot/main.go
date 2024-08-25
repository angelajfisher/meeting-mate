package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/angelajfisher/zoom-bot/internal/bot"
	"github.com/angelajfisher/zoom-bot/internal/server"
	"github.com/joho/godotenv"
	"github.com/oklog/run"
)

func main() {
	devMode, err := validateEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)

	g := run.Group{}

	g.Add(func() error {
		<-osSignal
		return nil
	}, func(error) {
		close(osSignal)
		err = bot.Stop()
		if err != nil {
			log.Println(err)
		}
	})

	g.Add(func() error { return server.Start(*devMode) }, func(error) {
		err = server.Stop()
		if err != nil {
			log.Println(err)
		}
	})

	err = bot.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: %s", err)
		os.Exit(1)
	}

	err = g.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: %s", err)
		os.Exit(1)
	}

	fmt.Println("See you later! o/")
}

func validateEnv() (*bool, error) {
	devMode := flag.Bool("dev", false, "run the program in development mode")
	envPath := flag.String("envFile", "", "program will load environment variables from the file at this path if provided")
	staticDir := flag.String("staticDir", "./static", "path to static directory containing site files")
	webhookPort := flag.String(
		"webhookPort",
		":12345",
		"port at which the webhook server will listen for incoming hooks - default :12345",
	)
	flag.Parse()

	if *envPath != "" {
		err := godotenv.Load(*envPath)
		if err != nil {
			return nil, errors.New("could not load .env file at provided path")
		}
	}

	if *devMode {
		fmt.Println("WARN: Initializing Zoom Bot in DEVELOPMENT mode")
	} else if os.Getenv("SSL_CERT") == "" || os.Getenv("SSL_KEY") == "" {
		return nil, errors.New("required SSL_CERT and/or SSL_KEY filepaths missing from environment")
	}

	server.BaseURL = "/projects/zoom-bot"
	server.StaticDir = *staticDir
	server.Port = *webhookPort
	server.Secret = os.Getenv("ZOOM_TOKEN")
	bot.BotToken = os.Getenv("BOT_TOKEN")
	bot.AppID = os.Getenv("APP_ID")

	if bot.BotToken == "" || bot.AppID == "" || server.Secret == "" {
		return nil, errors.New(
			"required ZOOM_TOKEN, BOT_TOKEN, and/or APP_ID variables missing from environment",
		)
	}

	return devMode, nil
}
