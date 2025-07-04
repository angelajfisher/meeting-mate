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

const (
	appVersion    = "1.4"
	fatalErrorMsg = "\nfatal: %v\n\nA fatal error occurred. Meeting Mate shut down.\n"
	separator     = "\n——————————————————————————————————————\n\n"
)

func Initialize(devMode bool) {
	fmt.Print(
		"\n┬┴┬┴┤･ω･)ﾉ├┬┴┬┴\n",
		"Hi, Welcome to Meeting Mate v"+appVersion+"!\n\n",
		"For help getting started, check out the documentation at:\n",
		"https://www.angelajfisher.com/projects/meeting-mate/docs\n",
	)

	botConfig, serverConfig, err := validateEnv(devMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, fatalErrorMsg, err)
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
		fmt.Fprintf(os.Stderr, fatalErrorMsg, err)
		os.Exit(1)
	}

	err = g.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, fatalErrorMsg, err)
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
	dbDisabled := flag.Bool(
		"dbDisabled",
		false,
		"disable the use of a sqlite3 database in favor of in-memory storage, which is lost on shutdown",
	)
	dbPathFlag := flag.String(
		"dbPath",
		"./.meetingmate-db.sqlite3",
		"preferred location of the database file",
	)
	sisterAddress := flag.String(
		"haURL",
		"",
		"URL of the other server in the high-availability pair",
	)
	flag.Parse()

	fmt.Println(separator + "Starting setup...\n\nLoading environment variables")
	defer fmt.Print(separator)

	if *envPath != "" {
		fmt.Println(
			"Loading variables from",
			*envPath,
			"— note that these will not override any existing environment variables",
		)
		err := godotenv.Load(*envPath)
		if err != nil {
			return nil, nil, errors.New("could not load .env file at provided path")
		}
	} else {
		fmt.Println("note: no .env file provided")
	}

	if *devMode {
		fmt.Print(
			"\nStarting in DEVELOPMENT mode:\n",
			"\t- Server running insecurely — HTTP without TLS\n",
			"\t- Zoom will NOT send webhook data in this mode\n",
			"\t- Meeting Mate WILL still connect to Discord\n",
			"This mode is for testing purposes only. The bot will not work as intended.\n",
		)
	} else if os.Getenv("SSL_CERT") == "" || os.Getenv("SSL_KEY") == "" {
		return nil, nil, errors.New("required SSL_CERT and/or SSL_KEY filepaths missing from environment")
	}

	var dbPool db.DatabasePool
	if *dbDisabled {
		fmt.Println("Database disabled — skipping initialization")
		dbPool = db.DatabasePool{Enabled: false}
	} else {
		var dbErr error
		dbPool, dbErr = setupDatabase(*dbPathFlag)
		if dbErr != nil {
			return nil, nil, fmt.Errorf("could not initialize database: %w", dbErr)
		}
	}

	if *sisterAddress == "" {
		fmt.Println("\nNo secondary address provided for high-availability — synchronization disabled")
	}

	o := orchestrator.NewOrchestrator(*sisterAddress, dbPool)
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
			"required variables ZOOM_TOKEN, BOT_TOKEN, and/or APP_ID missing from environment",
		)
	}

	fmt.Println("\nSetup complete! Time to get the party started!")

	return &botConf, &serverConf, nil
}

func setupDatabase(dbPath string) (db.DatabasePool, error) {
	cleanedDbPath := filepath.Clean(dbPath)
	fmt.Println("\nInitializing database at", cleanedDbPath)

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
