package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type internalErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Recovery(
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					recovered := recover()
					if recovered == nil {
						return
					}

					logger.ErrorContext(
						r.Context(),
						"panic recovered",
						"panic",
						recovered,
						"stack",
						string(debug.Stack()),
					)

					w.Header().Set(
						"Content-Type",
						"application/json; charset=utf-8",
					)
					w.WriteHeader(
						http.StatusInternalServerError,
					)

					err := json.NewEncoder(w).Encode(
						internalErrorResponse{
							Code:    "internal_error",
							Message: "internal server error",
						},
					)
					if err != nil {
						logger.ErrorContext(
							r.Context(),
							"failed to write recovery response",
							"error",
							err,
						)
					}
				}()

				next.ServeHTTP(w, r)
			},
		)
	}
}
