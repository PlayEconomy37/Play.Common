package opentelemetry

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTPMetrics is a struct that holds some prometheus metrics
// regarding HTTP requests
type HTTPMetrics struct {
	TotalRequestsCounter       *prometheus.CounterVec
	TotalResponsesCounter      *prometheus.CounterVec
	TotalProcessingTimeCounter *prometheus.HistogramVec
}

// CreateHTTPMetrics creates counters and histograms used to keep
// track of HTTP metrics in our application
func CreateHTTPMetrics(appName string) *HTTPMetrics {
	// Create total HTTP requests counter
	totalRequestsCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_total_requests_received", appName),
		Help: "Total HTTP requests received",
	}, []string{"method", "url"})

	// Create HTTP response counter
	totalResponsesCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_total_responses_sent", appName),
		Help: "Total HTTP responses sent",
	}, []string{"method", "url", "statusCode"})

	// Create HTTP requests duration histogram
	totalProcessingTimeCounter := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: fmt.Sprintf("%s_total_processing_time_microseconds", appName),
		Help: "Total processing time of HTTP requests in microseconds",
	}, []string{"method", "url"})

	return &HTTPMetrics{
		TotalRequestsCounter:       totalRequestsCounter,
		TotalResponsesCounter:      totalResponsesCounter,
		TotalProcessingTimeCounter: totalProcessingTimeCounter,
	}
}
