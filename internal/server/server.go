package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/angelajfisher/meeting-mate/internal/orchestrator"
)

type Config struct {
	DevMode      bool
	Orchestrator orchestrator.Orchestrator
	Port         string
	BaseURL      string
	StaticDir    string
	Secret       string
	server       *http.Server
	shuttingDown bool
}

const WEBHOOK_SLUG = "/webhooks/"

func Start(ss *Config) error {
	router := http.NewServeMux()
	fs := http.FileServer(http.Dir(ss.StaticDir))

	router.Handle("GET "+ss.BaseURL+"/static/", http.StripPrefix(ss.BaseURL+"/static/", fs))
	router.HandleFunc("GET "+ss.BaseURL+"/health", ss.handleHealth)
	router.HandleFunc("POST "+ss.BaseURL+WEBHOOK_SLUG, ss.handleWebhooks)
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

	log.Println("Starting Zoom webhook listener on", ss.Port)

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

// Health checks return 200 when the service is OK and 503 when shutting down
func (s *Config) handleHealth(w http.ResponseWriter, _ *http.Request) {
	select {
	case <-s.Orchestrator.ShutdownNotif:
		s.shuttingDown = true
	default:
	}
	if s.shuttingDown {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
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
