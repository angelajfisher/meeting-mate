package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/angelajfisher/zoom-bot/internal/types"
)

var (
	Port      string
	BaseURL   string
	StaticDir string
	meetingID string
)

func Start(devMode bool) {

	router := http.NewServeMux()
	fs := http.FileServer(http.Dir(StaticDir))

	router.Handle("GET "+BaseURL+"/static/", http.StripPrefix(BaseURL+"/static/", fs))
	router.HandleFunc("POST "+BaseURL+"/webhooks/", handleWebhooks)
	router.HandleFunc("GET "+BaseURL+"/", handleIndex)

	go func() {
		for {
			meetingID = <-types.WatchMeetingID

			if meetingID == types.Canceled {
				types.MeetingData <- types.EventData{EventType: types.Canceled}
				meetingID = ""
			}
		}
	}()

	log.Println("Zoom webhook listener starting on", Port)

	var err error
	if devMode {
		err = http.ListenAndServe(Port, router)
	} else {
		err = http.ListenAndServeTLS(Port, os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"), router)
	}
	if err != nil {
		log.Fatal("ERROR: Could not start Zoom webhook listener:", err)
	}

}

func handleIndex(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	http.ServeFile(w, r, filepath.Join(StaticDir, "/index.html"))

	elapsedTime := time.Since(startTime)
	formattedTime := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[%s] %s '%s' in %s\n", formattedTime, r.Method, r.URL.Path[len(BaseURL):], elapsedTime)

}
