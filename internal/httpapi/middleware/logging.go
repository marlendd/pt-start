package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}

	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(body []byte) (int, error) {
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

				writer := &loggingResponseWriter{
					ResponseWriter: w,
				}

				next.ServeHTTP(writer, r)

				status := writer.status
				if status == 0 {
					status = http.StatusOK
				}

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
