package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	TotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of get requests.",
		},
		[]string{"path"},
	)
	ProviderRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "provider_requests_total",
			Help: "Number of get requests to provider",
		},
		[]string{"provider"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(TotalRequests)
	prometheus.MustRegister(ProviderRequests)
}

func IncrementReqs(r *http.Request) {
	TotalRequests.WithLabelValues(r.URL.Path).Inc()
}

func IncrementProvider(provider string) {
	ProviderRequests.WithLabelValues(provider).Inc()
}
