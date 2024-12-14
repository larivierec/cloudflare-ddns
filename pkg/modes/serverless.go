package modes

import (
	"log"
	"net/http"
	"os"
	"sync"
)

type ServerlessMode struct {
	server *http.Server
	mu     sync.Mutex
}

func (m *ServerlessMode) Init() error {
	log.Println("Initializing Serverless Mode")
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create the HTTP server
	m.server = &http.Server{
		Addr:    ":9000",
		Handler: http.HandlerFunc(m.Invoke), // Use the Invoke function
	}

	return nil
}

func (m *ServerlessMode) Start() error {
	log.Println("Starting Serverless Mode")
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Serverless HTTP server failed: %v", err)
		}
	}()
	return nil
}

func (m *ServerlessMode) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.server != nil {
		log.Println("Stopping Serverless Mode")
		return m.server.Close()
	}

	return nil
}

// Invoke handles incoming requests in serverless mode.
func (m *ServerlessMode) Invoke(w http.ResponseWriter, r *http.Request) {
	createProvider()
	createCloudProvider()
	initialize()

	zoneName := os.Getenv("ZONE_NAME")
	recordName := os.Getenv("RECORD_NAME")

	rec, err := update(zoneName, recordName)
	if err != nil {
		log.Printf("DNS update failed: %v", err)
		http.Error(w, "DNS update failed", http.StatusInternalServerError)
		return
	}

	recordChecker(rec)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("DNS updated successfully"))
}
