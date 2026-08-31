package httpapi

import (
	"context"
	"net/http"
	"time"
)

const (
	readinessTimeout = time.Second
	healthStatusOK   = "ok"
)

func (h *Handler) Health(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.writeJSON(
		r.Context(),
		w,
		http.StatusOK,
		healthResponse{
			Status: healthStatusOK,
		},
	)
}

func (h *Handler) Ready(
	w http.ResponseWriter,
	r *http.Request,
) {
	if h.readiness == nil {
		h.writeNotReady(r.Context(), w)
		return
	}

	ctx, cancel := context.WithTimeout(
		r.Context(),
		readinessTimeout,
	)
	defer cancel()

	if err := h.readiness.Ping(ctx); err != nil {
		h.logger.DebugContext(
			r.Context(),
			"readiness check failed",
			"error",
			err,
		)

		h.writeNotReady(r.Context(), w)
		return
	}

	h.writeJSON(
		r.Context(),
		w,
		http.StatusOK,
		healthResponse{
			Status: healthStatusOK,
		},
	)
}

func (h *Handler) writeNotReady(
	ctx context.Context,
	w http.ResponseWriter,
) {
	h.writeError(
		ctx,
		w,
		http.StatusServiceUnavailable,
		errCodeNotReady,
		"service is not ready",
	)
}
