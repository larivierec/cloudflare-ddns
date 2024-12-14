package ddns

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/larivierec/cloudflare-ddns/pkg/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartHttpServer() {
	handlers.CloudProviderObj = cloudProviderObj
	handlers.ZoneName = zoneName
	handlers.RecordName = recordName

	health := new(handlers.HealthHandler)
	restart := new(handlers.RestartHandler)
	ddnsApi := new(handlers.ExternalHandler)

	healthRouter := http.NewServeMux()
	trafficRouter := http.NewServeMux()

	healthRouter.Handle("/metrics", promhttp.Handler())
	healthRouter.HandleFunc("/health/ready", health.Ready)
	healthRouter.HandleFunc("/health/alive", health.Alive)
	trafficRouter.HandleFunc("/v1/restart", restart.Do)
	trafficRouter.HandleFunc("/v1/get", ddnsApi.Get)
	trafficRouter.HandleFunc("/v1/set", ddnsApi.Set)

	healthServer := &http.Server{Addr: ":8080", Handler: healthRouter}
	trafficServer := &http.Server{Addr: ":9000", Handler: trafficRouter}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Health server error: %s", err)
		}
	}()

	go func() {
		if err := trafficServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Traffic server error: %s", err)
		}
	}()

	log.Println("HTTP servers started")
	<-done
	stopServers(healthServer, trafficServer)
}

func stopServers(healthServer, trafficServer *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := healthServer.Shutdown(ctx); err != nil {
		log.Printf("Health server shutdown error: %v", err)
	}
	if err := trafficServer.Shutdown(ctx); err != nil {
		log.Printf("Traffic server shutdown error: %v", err)
	}
	log.Println("Servers stopped gracefully")
}
