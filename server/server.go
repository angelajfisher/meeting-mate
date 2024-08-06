package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var Port string

func Start(devMode bool) {

	router := http.NewServeMux()

	router.HandleFunc("POST /projects/zoom-bot/webhooks", handleWebhooks)
	router.HandleFunc("GET /projects/zoom-bot", handleIndex)
	router.HandleFunc("GET /projects/zoom-bot/static/", serveStaticFiles)
	log.Println("Zoom webhook listener starting on", Port)

	var err error
	if devMode {
		err = http.ListenAndServe(Port, router)
	} else {
		err = http.ListenAndServeTLS(Port, os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"), router)
	}
	if err != nil {
		log.Fatal("Could not start Zoom webhook listener:", err)
	}

}

func handleIndex(w http.ResponseWriter, r *http.Request) {

	log.Println("Site accessed")

}

func serveStaticFiles(w http.ResponseWriter, r *http.Request) {

	filePath := r.URL.Path[len("projects/zoom-bot/static/"):]
	fullPath := filepath.Join(".", "static", filepath.Clean(filePath))
	http.ServeFile(w, r, fullPath)

}
