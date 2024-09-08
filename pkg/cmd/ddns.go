package ddns

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/cloudflare-go/v2/dns"
	"github.com/larivierec/cloudflare-ddns/pkg/api"
	"github.com/larivierec/cloudflare-ddns/pkg/ip"
	"github.com/larivierec/cloudflare-ddns/pkg/metrics"
	"github.com/larivierec/cloudflare-ddns/pkg/provider"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/spf13/pflag"
)

var (
	cachedIpInfo  = ""
	healthServer  = http.Server{}
	trafficServer = http.Server{}
	quit          = make(chan bool)
	done          = make(chan os.Signal, 1)
	providerName  = ""
	providers     = []api.Interface{}
	zoneName      string
	recordName    string
	isServerless  bool
	creds         api.CloudflareCredentials
)

type HealthHandler struct{}
type RestartHandler struct{}
type ExternalHandler struct{}

func (handle *HealthHandler) alive(w http.ResponseWriter, r *http.Request) {
	metrics.IncrementReqs(r)
	w.WriteHeader(http.StatusOK)
}

func (handle *HealthHandler) ready(w http.ResponseWriter, r *http.Request) {
	metrics.IncrementReqs(r)
	w.WriteHeader(http.StatusOK)
}

func (handle *RestartHandler) do(w http.ResponseWriter, r *http.Request) {
	metrics.IncrementReqs(r)
	w.WriteHeader(http.StatusAccepted)
	done <- syscall.SIGTERM
}

func (handle *ExternalHandler) get(w http.ResponseWriter, r *http.Request) {
	metrics.IncrementReqs(r)
	if cachedIpInfo == "" {
		cachedIpInfo, _ = provider.GetCurrentIP(*getProvider(providerName))
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(cachedIpInfo))
}

func initializeCloudflare() {
	creds = api.CloudflareCredentials{
		ApiKey:          os.Getenv("API_KEY"),
		AccountEmail:    os.Getenv("ACCOUNT_EMAIL"),
		CloudflareToken: os.Getenv("ACCOUNT_TOKEN"),
	}
	api.InitializeAPI(&creds)
}

func Start() {
	initializeCloudflare()
	pflag.StringVar(&zoneName, "zone-name", "", "set this to the cloudflare zone name.")
	pflag.StringVar(&recordName, "record-name", "", "set this to the cloudflare record in which you want to compare.")
	pflag.StringVar(&providerName, "provider", "ipify", "set this to the ip provider that will be queried for your public ip address.")
	pflag.Parse()

	createProvider()

	if !isServerless {
		metrics.InitMetrics()

		ticker := time.NewTicker(3 * time.Minute)
		go func() {
			for {
				select {
				case <-ticker.C:
					rec, err := update(zoneName, recordName)
					if err != nil {
						log.Println(err)
					}
					recordChecker(rec)
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()

		startHttpServer()
	} else {
		log.Println("running in serverless mode")
		rec, err := update(zoneName, recordName)
		if err != nil {
			log.Println(err)
		}
		recordChecker(rec)
	}
}

func startHttpServer() {
	health := new(HealthHandler)
	restart := new(RestartHandler)
	ddnsApi := new(ExternalHandler)
	healthRouter := http.NewServeMux()
	trafficRouter := http.NewServeMux()

	healthRouter.Handle("/metrics", promhttp.Handler())
	healthRouter.HandleFunc("/health/ready", health.ready)
	healthRouter.HandleFunc("/health/alive", health.alive)
	trafficRouter.HandleFunc("/v1/restart", restart.do)
	trafficRouter.HandleFunc("/v1/get", ddnsApi.get)

	healthServer.Handler = healthRouter
	healthServer.Addr = ":8080"
	trafficServer.Handler = trafficRouter
	trafficServer.Addr = ":9000"

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen health server: %s\n", err)
		}
	}()

	go func() {
		if err := trafficServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen traffic server: %s\n", err)
		}
	}()

	log.Println("server started")
	<-done
	stopServer()
}

func StartServerless(w http.ResponseWriter, r *http.Request) {
	initializeCloudflare()
	zoneName := os.Getenv("ZONE_NAME")
	recordName := os.Getenv("RECORD_NAME")

	createProvider()
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

func update(zoneName string, recordName string) (*dns.Record, error) {
	var rec *dns.Record
	result, err := provider.GetCurrentIP(*getProvider(providerName))
	if err != nil {
		return nil, fmt.Errorf("unable to get provider ip, skipping update. error: %v", err)
	}
	if cachedIpInfo != result {
		cachedIpInfo = result
		record, zoneId, err := api.ListDNSRecordsFiltered(zoneName, recordName)
		if err != nil {
			return nil, fmt.Errorf("unable to filter for %s. err: %s", recordName, err)
		}

		if cachedIpInfo != record.Content {
			rec, err = api.UpdateDNSRecord(result, zoneId, record)
			if err != nil {
				return nil, fmt.Errorf("unable to update record %s. err : %s", recordName, err)
			}
		}
	} else {
		log.Println("IPs are the same")
	}
	return rec, nil
}

func stopServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer func() {
		cancel()
	}()

	if err := healthServer.Shutdown(ctx); err != nil {
		log.Fatalf("health server unable to shutdown: %v", err)
	}

	if err := trafficServer.Shutdown(ctx); err != nil {
		log.Fatalf("traffic server unable to shutdown: %v", err)
	}
	log.Println("servers stopped gracefully")
}

func createProvider() {
	switch providerName {
	case "ipify":
		providers = append(providers, &ip.Ipify{})
	case "icanhazip":
	case "icanhaz":
		providers = append(providers, &ip.ICanHazIp{})
	case "random":
		providers = append(providers, &ip.Ipify{})
		providers = append(providers, &ip.ICanHazIp{})
	default:
		providerName = "random"
		providers = append(providers, &ip.Ipify{})
		providers = append(providers, &ip.ICanHazIp{})
	}
}

func getProvider(typeName string) *api.Interface {
	if typeName != "random" {
		return &providers[0]
	} else {
		return &providers[rand.Intn(len(providers))]
	}
}

func recordChecker(record *dns.Record) {
	if record != nil {
		log.Printf("record updated to: %s\n", record.Content)
	} else {
		log.Println("record is the same, ignoring.")
	}
}
