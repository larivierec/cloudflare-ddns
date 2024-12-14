package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/larivierec/cloudflare-ddns/pkg/cloudprovider"
	"github.com/larivierec/cloudflare-ddns/pkg/metrics"
)

var (
	cachedIpInfo     = ""
	CloudProviderObj cloudprovider.Provider
	ZoneName         string
	RecordName       string
)

type HealthHandler struct{}
type RestartHandler struct{}
type ExternalHandler struct{}

func (handle *HealthHandler) Alive(w http.ResponseWriter, r *http.Request) {
	metrics.IncrementReqs(r)
	w.WriteHeader(http.StatusOK)
}

func (handle *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	metrics.IncrementReqs(r)
	w.WriteHeader(http.StatusOK)
}

func (handle *RestartHandler) Do(w http.ResponseWriter, r *http.Request) {
	metrics.IncrementReqs(r)
	w.WriteHeader(http.StatusAccepted)
	// Add logic to trigger a graceful shutdown
}

func (handle *ExternalHandler) Get(w http.ResponseWriter, r *http.Request) {
	metrics.IncrementReqs(r)
	if cachedIpInfo == "" {
		// Obtain the IP information (this depends on your core logic)
		log.Println("Fetching cached IP info...")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(cachedIpInfo))
}

func (handle *ExternalHandler) Set(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, "Missing 'ip' parameter", http.StatusBadRequest)
		return
	}
	newRecord := &cloudprovider.Record{}
	current, _ := CloudProviderObj.GetDNSRecord(ZoneName, RecordName)
	CloudProviderObj.FillRecord(current, newRecord)
	newRecord.Content = ip
	if current == nil {
		fmt.Printf("Record %s doesn't exist, creating.\n", RecordName)
		CloudProviderObj.InitializeRecord(ZoneName, *newRecord)
	}
	responseRecord, _ := CloudProviderObj.UpdateDNSRecord(ZoneName, *newRecord)
	responseRecBytes, _ := json.Marshal(responseRecord)
	metrics.IncrementReqs(r)
	w.WriteHeader(http.StatusOK)
	w.Write(responseRecBytes)
}
