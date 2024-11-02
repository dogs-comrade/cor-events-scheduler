// internal/metrics/metrics.go
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once sync.Once

	HttpRequestsTotal    *prometheus.CounterVec
	HttpRequestDuration  *prometheus.HistogramVec
	HttpRequestsInFlight prometheus.Gauge
	HttpRequestSize      *prometheus.HistogramVec
	HttpResponseSize     *prometheus.HistogramVec
)

// InitMetrics инициализирует все метрики
func InitMetrics() {
	once.Do(func() {
		HttpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "events_scheduler",
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests by method and path",
			},
			[]string{"method", "path", "status"},
		)

		HttpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "events_scheduler",
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration distribution",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		)

		HttpRequestsInFlight = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "events_scheduler",
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		)

		HttpRequestSize = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "events_scheduler",
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		)

		HttpResponseSize = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "events_scheduler",
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path", "status"},
		)

		// Регистрируем все метрики
		prometheus.MustRegister(
			HttpRequestsTotal,
			HttpRequestDuration,
			HttpRequestsInFlight,
			HttpRequestSize,
			HttpResponseSize,
		)
	})
}
