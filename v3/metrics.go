package luddite

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests made.",
	}, []string{"code", "method"})

	httpRequestDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Subsystem:  "http",
		Name:       "request_duration_seconds",
		Help:       "The HTTP request latencies in seconds.",
		Objectives: summaryObjectives,
	}, []string{"method"})

	httpRequestsInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "http",
		Name:      "requests_in_flight",
		Help:      "Current number of HTTP requests in flight.",
	})

	httpRequestSizeBytes = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Subsystem:  "http",
		Name:       "request_size_bytes",
		Help:       "The HTTP response sizes in bytes.",
		Objectives: summaryObjectives,
	}, []string{"method"})

	httpResponseSizeBytes = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Subsystem:  "http",
		Name:       "response_size_bytes",
		Help:       "The HTTP response sizes in bytes.",
		Objectives: summaryObjectives,
	}, []string{"method"})

	summaryObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
)

func init() {
	_ = prometheus.Register(httpRequestsTotal)
	_ = prometheus.Register(httpRequestDuration)
	_ = prometheus.Register(httpRequestsInFlight)
	_ = prometheus.Register(httpRequestSizeBytes)
	_ = prometheus.Register(httpResponseSizeBytes)
}

func instrumentHTTPHandler(h http.Handler) http.Handler {
	return promhttp.InstrumentHandlerInFlight(httpRequestsInFlight,
		promhttp.InstrumentHandlerDuration(httpRequestDuration,
			promhttp.InstrumentHandlerCounter(httpRequestsTotal,
				promhttp.InstrumentHandlerRequestSize(httpRequestSizeBytes,
					promhttp.InstrumentHandlerResponseSize(httpResponseSizeBytes, h)))))
}
