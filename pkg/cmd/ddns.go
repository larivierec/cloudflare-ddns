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

	"github.com/larivierec/cloudflare-ddns/pkg/cloudprovider"
	"github.com/larivierec/cloudflare-ddns/pkg/cloudprovider/cloudflare"
	"github.com/larivierec/cloudflare-ddns/pkg/ipprovider"
	"github.com/larivierec/cloudflare-ddns/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/spf13/pflag"
)

var (
	cachedIpInfo           = ""
	healthServer           = http.Server{}
	trafficServer          = http.Server{}
	quit                   = make(chan bool)
	done                   = make(chan os.Signal, 1)
	ipProviderName         = ""
	ipProviders            = []ipprovider.Provider{}
	requestedCloudProvider string
	zoneName               string
	recordName             string
	isServerless           bool
	ticker                 time.Duration
	cloudProviderObj       cloudprovider.Provider
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
		provider := *getProvider(ipProviderName)
		cachedIpInfo, _ = ipprovider.GetCurrentIP(provider, metrics.IncrementProvider)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(cachedIpInfo))
}

func Start() {
	pflag.StringVar(&requestedCloudProvider, "cloud-provider", "cloudflare", "set this to the requested cloud provider. where your `A` record will be created.")
	pflag.StringVar(&zoneName, "zone-name", "", "set this to the cloudflare zone name.")
	pflag.StringVar(&recordName, "record-name", "", "set this to the cloudflare record in which you want to compare.")
	pflag.StringVar(&ipProviderName, "provider", "ipify", "set this to the ip provider that will be queried for your public ip address.")
	pflag.DurationVar(&ticker, "ticker", time.Duration(3*time.Minute), "set this to the desired time to check your WAN IP against the IP Providers.")
	pflag.Parse()

	createProvider()
	createCloudProvider()

	if !isServerless {
		metrics.InitMetrics()

		ticker := time.NewTicker(ticker)
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
	createProvider()
	createCloudProvider()
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

func update(zoneName string, recordName string) (*cloudprovider.Record, error) {
	var rec = &cloudprovider.Record{}
	result, err := ipprovider.GetCurrentIP(*getProvider(ipProviderName), metrics.IncrementProvider)
	if err != nil {
		return nil, fmt.Errorf("unable to get provider ip, skipping update. error: %v", err)
	}
	if cachedIpInfo != result {
		cachedIpInfo = result
		record, err := cloudProviderObj.GetDNSRecord(zoneName, recordName)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve record %s in zone %s. err: %s", recordName, zoneName, err)
		}
		cloudProviderObj.FillRecord(record, rec)
		if cachedIpInfo != record["content"] {
			_, err := cloudProviderObj.UpdateDNSRecord(zoneName, *rec)
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
	switch ipProviderName {
	case "ipify":
		ipProviders = append(ipProviders, &ipprovider.Ipify{})
	case "icanhazip":
	case "icanhaz":
		ipProviders = append(ipProviders, &ipprovider.ICanHazIp{})
	case "random":
		ipProviders = append(ipProviders, &ipprovider.Ipify{})
		ipProviders = append(ipProviders, &ipprovider.ICanHazIp{})
	default:
		ipProviderName = "random"
		ipProviders = append(ipProviders, &ipprovider.Ipify{})
		ipProviders = append(ipProviders, &ipprovider.ICanHazIp{})
	}
}

func createCloudProvider() {
	switch requestedCloudProvider {
	case "cloudflare":
		cloudProviderObj = cloudflare.NewCloudflareProvider()
	default:
		cloudProviderObj = cloudflare.NewCloudflareProvider()
	}
}

func getProvider(typeName string) *ipprovider.Provider {
	if typeName != "random" {
		return &ipProviders[0]
	} else {
		return &ipProviders[rand.Intn(len(ipProviders))]
	}
}

func recordChecker(record *cloudprovider.Record) {
	if cachedIpInfo != record.Content {
		log.Printf("record updated to: %s\n", record.Content)
	} else {
		log.Println("record is the same, ignoring.")
	}
}
