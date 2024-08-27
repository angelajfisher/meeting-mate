package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/angelajfisher/meeting-mate/internal/types"
)

var (
	Port      string
	BaseURL   string
	StaticDir string
	meetingID string
	server    *http.Server
)

func Start(devMode bool) error {
	router := http.NewServeMux()
	fs := http.FileServer(http.Dir(StaticDir))

	router.Handle("GET "+BaseURL+"/static/", http.StripPrefix(BaseURL+"/static/", fs))
	router.HandleFunc("POST "+BaseURL+"/webhooks/", handleWebhooks)
	router.HandleFunc("GET "+BaseURL+"/docs", handleDocs)
	router.HandleFunc("GET "+BaseURL+"/", handleIndex)

	server = &http.Server{
		Addr:              Port,
		Handler:           http.TimeoutHandler(router, 5*time.Second, "Oops, timed out!"),
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	go func() {
		for {
			meetingID = <-types.WatchMeetingID

			switch meetingID {
			case types.Canceled:
				meetingID = ""
				types.MeetingData <- types.EventData{EventType: types.Canceled}
			case types.Shutdown:
				meetingID = ""
				types.MeetingData <- types.EventData{EventType: types.Shutdown}
			}
		}
	}()

	log.Println("Zoom webhook listener starting on", Port)

	var err error
	if devMode {
		err = server.ListenAndServe()
	} else {
		err = server.ListenAndServeTLS(os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"))
	}
	if err != nil {
		return fmt.Errorf("could not start Zoom webhook listener: %w", err)
	}

	return nil
}

func Stop() error {
	if server == nil {
		return nil
	}

	fmt.Print("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("could not shutdown server gracefully: %w", err)
	}

	fmt.Print("Done!\n")
	return nil
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	http.ServeFile(w, r, filepath.Join(StaticDir, "/index.html"))

	elapsedTime := time.Since(startTime)
	log.Printf("%s '%s' in %s\n", r.Method, r.URL.Path[len(BaseURL):], elapsedTime)
}

func handleDocs(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	http.ServeFile(w, r, filepath.Join(StaticDir, "/docs.html"))

	elapsedTime := time.Since(startTime)
	log.Printf("%s '%s' in %s\n", r.Method, r.URL.Path[len(BaseURL):], elapsedTime)
}
