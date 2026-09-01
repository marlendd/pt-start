package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}

	return w.status
}

func (w *responseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}

	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}

	size, err := w.ResponseWriter.Write(body)
	w.size += size

	return size, err
}

func RequestLogging(
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				startedAt := time.Now()

				writer := &responseWriter{
					ResponseWriter: w,
				}

				next.ServeHTTP(writer, r)

				status := writer.Status()

				logger.InfoContext(
					r.Context(),
					"HTTP request completed",
					"method",
					r.Method,
					"path",
					r.URL.Path,
					"status",
					status,
					"response_bytes",
					writer.size,
					"duration",
					time.Since(startedAt),
					"remote_address",
					r.RemoteAddr,
				)
			},
		)
	}
}
