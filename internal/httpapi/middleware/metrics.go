package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const unknownRoute = "unknown"

type HTTPMetrics struct {
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
}

func NewHTTPMetrics(
	registerer prometheus.Registerer,
) *HTTPMetrics {
	metrics := &HTTPMetrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "shortener",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests.",
			},
			[]string{"method", "route", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "shortener",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "route"},
		),
		requestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "shortener",
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Number of HTTP requests currently running.",
			},
		),
	}

	registerer.MustRegister(
		metrics.requestsTotal,
		metrics.requestDuration,
		metrics.requestsInFlight,
	)

	return metrics
}

func (m *HTTPMetrics) Middleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			m.requestsInFlight.Inc()

			writer := &responseWriter{
				ResponseWriter: w,
			}

			defer func() {
				m.requestsInFlight.Dec()

				route := r.Pattern
				if route == "" {
					route = unknownRoute
				}

				m.requestsTotal.WithLabelValues(
					r.Method,
					route,
					strconv.Itoa(writer.Status()),
				).Inc()

				m.requestDuration.WithLabelValues(
					r.Method,
					route,
				).Observe(
					time.Since(startedAt).Seconds(),
				)
			}()

			next.ServeHTTP(writer, r)

		},
	)
}
