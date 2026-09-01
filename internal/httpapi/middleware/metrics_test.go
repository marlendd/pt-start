package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestHTTPMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewHTTPMetrics(registry)

	mux := http.NewServeMux()
	mux.HandleFunc(
		"/shorten",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		},
	)

	handler := metrics.Middleware(mux)

	request := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code)

	require.Equal(
		t,
		float64(1),
		testutil.ToFloat64(
			metrics.requestsTotal.WithLabelValues(
				http.MethodPost,
				"/shorten",
				"201",
			),
		),
	)

	require.Zero(
		t,
		testutil.ToFloat64(metrics.requestsInFlight),
	)

	require.Equal(
		t,
		1,
		testutil.CollectAndCount(
			metrics.requestDuration,
			"shortener_http_request_duration_seconds",
		),
	)
}
