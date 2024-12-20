package ddns

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/larivierec/cloudflare-ddns/pkg/cloudprovider"
	"github.com/larivierec/cloudflare-ddns/pkg/cloudprovider/cloudflare"
	"github.com/larivierec/cloudflare-ddns/pkg/ipprovider"
	"github.com/larivierec/cloudflare-ddns/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/spf13/pflag"
)

type applicationMode int

const (
	application applicationMode = iota
	serverless
	api
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
	mode                   applicationMode
	ticker                 time.Duration
	createMissing          bool
	recordTTL              int
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

// external update, i.e. unifi gw
func (handle *ExternalHandler) set(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, "Missing 'ip' parameter", http.StatusBadRequest)
		return
	}
	newRecord := &cloudprovider.Record{}
	current, _ := cloudProviderObj.GetDNSRecord(zoneName, recordName)
	cloudProviderObj.FillRecord(current, newRecord)
	newRecord.Content = ip
	if current == nil {
		fmt.Printf("record %s, doesn't exist in cloud provider, creating.", recordName)
		cloudProviderObj.InitializeRecord(zoneName, *newRecord)
	}
	responseRecord, _ := cloudProviderObj.UpdateDNSRecord(zoneName, *newRecord)
	responseRecBytes, _ := json.Marshal(responseRecord)
	metrics.IncrementReqs(r)
	w.WriteHeader(http.StatusOK)
	w.Write(responseRecBytes)
}

func Start(chosenMode string) {
	modeSelector(chosenMode)
	pflag.StringVar(&requestedCloudProvider, "cloud-provider", "cloudflare", "set this to the requested cloud provider. where your `A` record will be created")
	pflag.StringVar(&zoneName, "zone-name", "", "set this to the cloudflare zone name")
	pflag.StringVar(&recordName, "record-name", "", "set this to the cloudflare record in which you want to compare")
	pflag.StringVar(&ipProviderName, "provider", "ipify", "set this to the ip provider that will be queried for your public ip address")
	pflag.DurationVar(&ticker, "ticker", time.Duration(3*time.Minute), "set this to the desired time to check your WAN IP against the IP Providers")
	pflag.BoolVar(&createMissing, "create-missing", false, "create missing ddns record for updating")
	pflag.IntVar(&recordTTL, "record-ttl", 300, "set this to the value of the requested TTL")
	pflag.Parse()

	createProvider()
	createCloudProvider()
	initialize()

	if mode != serverless {
		metrics.InitMetrics()

		if mode == application {
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
		} else {
			go func() {
				for {
					select {
					case <-quit:
						return
					}
				}
			}()
		}
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
	trafficRouter.HandleFunc("/v1/set", ddnsApi.set)

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
	ipProviderName = strings.ToLower(strings.TrimSpace(ipProviderName))
	switch ipProviderName {
	case "ipify":
		ipProviders = append(ipProviders, &ipprovider.Ipify{})
	case "icanhazip", "icanhaz":
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

func initialize() {
	result, err := ipprovider.GetCurrentIP(*getProvider(ipProviderName), metrics.IncrementProvider)
	if err != nil {
		log.Fatalf("unable to initialize retrieve ip for init, aborting.")
	}
	providerRecord, _ := cloudProviderObj.GetDNSRecord(zoneName, recordName)
	if createMissing && providerRecord == nil {
		log.Printf("creating %s, in zone %s", recordName, zoneName)
		rec := &cloudprovider.Record{
			Type:    "A",
			TTL:     recordTTL,
			Name:    recordName,
			Content: result,
		}
		cloudProviderObj.InitializeRecord(zoneName, *rec)
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
	if record != nil {
		if cachedIpInfo != record.Content && record.Content != "" {
			log.Printf("record updated to: %s\n", record.Content)
		} else {
			log.Println("record is the same, ignoring.")
		}
	}
}

func modeSelector(chosenMode string) {
	switch chosenMode {
	case "api":
		mode = api
	case "serverless":
		mode = serverless
	case "application":
		mode = application
	default:
		mode = application
	}
}
