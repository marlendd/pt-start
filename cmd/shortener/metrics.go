package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/marlendd/pt-start/internal/httpapi/middleware"
)

func newMetrics() (
	*middleware.HTTPMetrics,
	http.Handler,
) {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{},
		),
	)

	httpMetrics := middleware.NewHTTPMetrics(registry)

	metricsHandler := promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{},
	)

	return httpMetrics, metricsHandler
}
