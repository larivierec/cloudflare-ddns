package metrics

import (
	"net/http"

	"github.com/larivierec/cloudflare-ddns/pkg/api"
	"github.com/prometheus/client_golang/prometheus"
)

var TotalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"path"},
)

var ProviderRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "provider_requests_total",
		Help: "Number of get requests to provider",
	},
	[]string{"provider"},
)

func InitMetrics() {
	prometheus.Register(TotalRequests)
	prometheus.Register(ProviderRequests)
}

func IncrementProvider(provider api.Interface) {
	ProviderRequests.WithLabelValues(provider.GetProviderName()).Inc()
}

func IncrementReqs(r *http.Request) {
	TotalRequests.WithLabelValues(r.URL.Path).Inc()
}
