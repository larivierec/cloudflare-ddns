package ddns

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/larivierec/cloudflare-ddns/pkg/api"
	"github.com/larivierec/cloudflare-ddns/pkg/ipify"
	"github.com/spf13/pflag"
	"github.com/thecodeteam/goodbye"
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
				update(zoneName, recordName)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	log.Println("Starting http server")
	startHttpServer()
}

func startHttpServer() {
	health := new(HealthHandler)
	mainRouter := mux.NewRouter()
	healthRouter := mainRouter.PathPrefix("/health").Subrouter()
	healthRouter.HandleFunc("/ready", health.ready).Methods(http.MethodGet)
	healthRouter.HandleFunc("/alive", health.alive).Methods(http.MethodGet)
	server.Handler = mainRouter
	listener, err := net.Listen("tcp", "0.0.0.0:8080")

	if err != nil {
		log.Fatalf("unable to start http server %v", err.Error())
	}

	err = server.Serve(listener)
	if err != nil {
		log.Fatalf("unable to serve %v", err.Error())
	}

	WaitForCtrlC()
}

func update(zoneName string, recordName string) {
	ipifyResult, err := ipify.GetCurrentIP()
	if err != nil {
		log.Fatalf("Unable to get ipify ip, aborting.")
	}
	if cachedIpInfo.Ip != ipifyResult.Ip {
		cachedIpInfo.Ip = ipifyResult.Ip
		record, zoneId, err := api.ListDNSRecordsFiltered(zoneName, recordName)
		if err != nil {
			log.Println(fmt.Errorf("unable to filter for %s. err: %s", recordName, err))
		}

		if cachedIpInfo.Ip != record.Content {
			err = api.UpdateDNSRecord(ipifyResult.Ip, zoneId, record)
			if err != nil {
				log.Println(fmt.Errorf("unable to update record %s. err : %s", recordName, err))
			}
		}
	} else {
		log.Println("IPs are the same")
	}
}

func WaitForCtrlC() {
	done := make(chan bool, 1)

	goodbye.RegisterWithPriority(func(ctx context.Context, s os.Signal) {
		server.Shutdown(ctx)
		done <- true
	}, 999)

	<-done
}
