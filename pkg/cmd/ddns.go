package ddns

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/larivierec/cloudflare-ddns/pkg/api"
	"github.com/larivierec/cloudflare-ddns/pkg/ipify"
	"github.com/spf13/pflag"
)

var (
	cachedIpInfo ipify.IpInfo
	server       = http.Server{}
)

type HealthHandler struct{}
type RestartHandler struct{}

func (handle *HealthHandler) alive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (handle *HealthHandler) ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Start() {
	creds := api.CloudflareCredentials{
		ApiKey:          os.Getenv("API_KEY"),
		AccountEmail:    os.Getenv("ACCOUNT_EMAIL"),
		CloudflareToken: os.Getenv("ACCOUNT_TOKEN"),
	}

	var zoneName string
	var recordName string

	pflag.StringVar(&zoneName, "zone-name", "", "set this to the cloudflare zone name")
	pflag.StringVar(&recordName, "record-name", "", "set this to the cloudflare record in which you want to compare")
	pflag.Parse()

	err := api.InitializeAPI(&creds)

	if err != nil {
		log.Fatalf("unable to initialize cloudflare api")
	}

	ticker := time.NewTicker(3 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				err := update(zoneName, recordName)
				if err != nil {
					log.Println(err)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	startHttpServer()
}

func startHttpServer() {
	health := new(HealthHandler)
	mainRouter := http.NewServeMux()
	mainRouter.HandleFunc("/health/ready", health.ready)
	mainRouter.HandleFunc("/health/alive", health.alive)
	server.Handler = mainRouter
	server.Addr = ":8080"

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("server started")
	<-done
	stopServer()
}

func update(zoneName string, recordName string) error {
	ipifyResult, err := ipify.GetCurrentIP()
	if err != nil {
		return fmt.Errorf("unable to get ipify ip, skipping update. error: %v", err)
	}
	if cachedIpInfo.Ip != ipifyResult.Ip {
		cachedIpInfo.Ip = ipifyResult.Ip
		record, zoneId, err := api.ListDNSRecordsFiltered(zoneName, recordName)
		if err != nil {
			return fmt.Errorf("unable to filter for %s. err: %s", recordName, err)
		}

		if cachedIpInfo.Ip != record.Content {
			err = api.UpdateDNSRecord(ipifyResult.Ip, zoneId, record)
			if err != nil {
				return fmt.Errorf("unable to update record %s. err : %s", recordName, err)
			}
		}
	} else {
		log.Println("IPs are the same")
	}
	return nil
}

func stopServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server unable to shutdown: %v", err)
	}
	log.Println("server stopped gracefully")
}
