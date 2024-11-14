package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	DevMode   bool
	Port      string
	BaseURL   string
	StaticDir string
	Secret    string
	server    *http.Server
}

func Start(ss *Config) error {
	router := http.NewServeMux()
	fs := http.FileServer(http.Dir(ss.StaticDir))

	router.Handle("GET "+ss.BaseURL+"/static/", http.StripPrefix(ss.BaseURL+"/static/", fs))
	router.HandleFunc("POST "+ss.BaseURL+"/webhooks/", ss.handleWebhooks)
	router.HandleFunc("GET "+ss.BaseURL+"/docs", ss.handleDocs)
	router.HandleFunc("GET "+ss.BaseURL+"/", ss.handleIndex)

	ss.server = &http.Server{
		Addr:              ss.Port,
		Handler:           http.TimeoutHandler(router, 5*time.Second, "Oops, timed out!"),
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Println("Zoom webhook listener starting on", ss.Port)

	var err error
	if ss.DevMode {
		err = ss.server.ListenAndServe()
	} else {
		err = ss.server.ListenAndServeTLS(os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"))
	}
	if err != nil {
		return fmt.Errorf("could not start Zoom webhook listener: %w", err)
	}

	return nil
}

func Stop(ss *Config) error {
	if ss.server == nil {
		return nil
	}

	fmt.Print("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ss.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("could not shutdown server gracefully: %w", err)
	}

	fmt.Print("Done!\n")
	return nil
}

func (s Config) handleIndex(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	http.ServeFile(w, r, filepath.Join(s.StaticDir, "/index.html"))

	elapsedTime := time.Since(startTime)
	log.Printf("%s '%s' in %s\n", r.Method, r.URL.Path[len(s.BaseURL):], elapsedTime)
}

func (s Config) handleDocs(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	http.ServeFile(w, r, filepath.Join(s.StaticDir, "/docs.html"))

	elapsedTime := time.Since(startTime)
	log.Printf("%s '%s' in %s\n", r.Method, r.URL.Path[len(s.BaseURL):], elapsedTime)
}
