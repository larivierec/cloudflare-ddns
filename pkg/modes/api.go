package modes

import (
	"log"
	"net/http"
	"os"

	"github.com/larivierec/cloudflare-ddns/pkg/handlers"
)

type APIMode struct {
	server *http.Server
	done   chan os.Signal
}

func (m *APIMode) Init() error {
	log.Println("Initializing API Mode")
	m.done = make(chan os.Signal, 1)
	return nil
}

func (m *APIMode) Start() error {
	log.Println("Starting API Mode")

	mux := http.NewServeMux()
	healthHandler := &handlers.HealthHandler{}
	restartHandler := &handlers.RestartHandler{}
	externalHandler := &handlers.ExternalHandler{}

	mux.HandleFunc("/health/alive", healthHandler.Alive)
	mux.HandleFunc("/health/ready", healthHandler.Ready)
	mux.HandleFunc("/restart", restartHandler.Do)
	mux.HandleFunc("/external/get", externalHandler.Get)
	mux.HandleFunc("/external/set", externalHandler.Set)

	m.server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API HTTP server failed: %v", err)
		}
	}()

	return nil
}

func (m *APIMode) Stop() error {
	if m.server != nil {
		log.Println("Stopping API Mode")
		return m.server.Close()
	}
	return nil
}
