package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var Port string
var BaseURL string

func Start(devMode bool) {

	router := http.NewServeMux()
	// fs := http.FileServer(http.Dir("./static"))

	// router.Handle("GET "+BaseURL+"/static/", http.StripPrefix(BaseURL+"/static/", fs))
	router.HandleFunc("POST "+BaseURL+"/webhooks", handleWebhooks)
	router.HandleFunc("GET /", handleStaticSite)

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

func handleStaticSite(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	if len(r.URL.Path) < len(BaseURL) {
		http.NotFound(w, r)
		log.Printf("Could not serve path %v\n", r.URL.Path)
		return
	}

	fullPath := filepath.Join(".", "static", filepath.Clean(r.URL.Path[len(BaseURL):])+".html")
	if fullPath == "static/.html" {
		fullPath = "static/index.html"
	}

	log.Printf("Looking for file with path %v\n", fullPath)

	// 404 for nonexistent files
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			log.Printf("Could not find file at %v: %s\n", fullPath, err)
			return
		}
	}
	if info.IsDir() {
		http.NotFound(w, r)
		log.Printf("Could not find file at %v: %s\n", fullPath, err)
		return
	}

	http.ServeFile(w, r, fullPath)

	elapsedTime := time.Since(startTime)
	formattedTime := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[%s] %s '%s' in %s\n", formattedTime, r.Method, r.URL.Path, elapsedTime)

}
