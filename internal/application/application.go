// Boilerplate for initializing the program
package application

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/angelajfisher/meeting-mate/internal/bot"
	"github.com/angelajfisher/meeting-mate/internal/db"
	"github.com/angelajfisher/meeting-mate/internal/orchestrator"
	"github.com/angelajfisher/meeting-mate/internal/server"
	"github.com/joho/godotenv"
	"github.com/oklog/run"
)

func Initialize(devMode bool) {
	botConfig, serverConfig, err := validateEnv(devMode)
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
		err = bot.Stop(botConfig)
		if err != nil {
			log.Println(err)
		}
	})

	g.Add(func() error { return server.Start(serverConfig) }, func(error) {
		err = server.Stop(serverConfig)
		if err != nil {
			log.Println(err)
		}
	})

	err = bot.Run(botConfig)
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

func validateEnv(dev bool) (*bot.Config, *server.Config, error) {
	devMode := flag.Bool("dev", dev, "run the program in development mode")
	envPath := flag.String("envFile", "", "program will load environment variables from the file at this path if provided")
	staticDir := flag.String("staticDir", "./static", "path to static directory containing site files")
	webhookPort := flag.String(
		"webhookPort",
		":12345",
		"port at which the webhook server will listen for incoming hooks",
	)
	dbPathFlag := flag.String(
		"dbPath",
		"./.meetingmate-db.sqlite3",
		"preferred location of the database file",
	)
	flag.Parse()

	if *envPath != "" {
		err := godotenv.Load(*envPath)
		if err != nil {
			return nil, nil, errors.New("could not load .env file at provided path")
		}
	}

	if *devMode {
		fmt.Println("WARN: Initializing Meeting Mate in DEVELOPMENT mode")
	} else if os.Getenv("SSL_CERT") == "" || os.Getenv("SSL_KEY") == "" {
		return nil, nil, errors.New("required SSL_CERT and/or SSL_KEY filepaths missing from environment") //todo: suggest env file
	}

	dbPool, err := setupDatabase(*dbPathFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("could not initialize database: %w", err)
	}

	o := orchestrator.NewOrchestrator(dbPool)
	botConf := bot.Config{
		BotToken:     os.Getenv("BOT_TOKEN"),
		AppID:        os.Getenv("APP_ID"),
		Orchestrator: o,
	}
	serverConf := server.Config{
		DevMode:      *devMode,
		Orchestrator: o,
		BaseURL:      "/projects/meeting-mate",
		StaticDir:    *staticDir,
		Port:         *webhookPort,
		Secret:       os.Getenv("ZOOM_TOKEN"),
	}

	if botConf.BotToken == "" || botConf.AppID == "" || serverConf.Secret == "" {
		return nil, nil, errors.New(
			"required ZOOM_TOKEN, BOT_TOKEN, and/or APP_ID variables missing from environment",
		)
	}

	return &botConf, &serverConf, nil
}

func setupDatabase(dbPath string) (db.DatabasePool, error) {
	cleanedDbPath := filepath.Clean(dbPath)
	log.Println("Initializing database at", cleanedDbPath)

	err := db.InitializeDatabase(cleanedDbPath)
	if err != nil {
		return db.DatabasePool{}, fmt.Errorf("could not create database: %w", err)
	}
	err = db.MakeMigrations(cleanedDbPath)
	if err != nil {
		return db.DatabasePool{}, fmt.Errorf("could not make database migrations: %w", err)
	}

	dbPool, err := db.NewDatabasePool(cleanedDbPath)
	if err != nil {
		err = fmt.Errorf("could not create database pool: %w", err)
	}

	return dbPool, err
}
