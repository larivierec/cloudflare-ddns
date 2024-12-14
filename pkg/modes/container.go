package modes

import (
	"log"
	"time"
)

type Container struct {
	ticker *time.Ticker
}

func (m *Container) Init() error {
	log.Println("Initializing Application Mode")
	m.ticker = time.NewTicker(3 * time.Minute)
	return nil
}

func (m *Container) Start() error {
	log.Println("Starting Application Mode")
	go func() {
		for range m.ticker.C {
			log.Println("Application Mode ticker triggered")
			// Add update logic here
		}
	}()
	return nil
}

func (m *Container) Stop() error {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	log.Println("Stopping Application Mode")
	return nil
}
