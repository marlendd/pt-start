package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
)

func (h *Handler) writeJSON(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	payload any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.WarnContext(
			ctx,
			"failed to write JSON response",
			"error",
			err,
		)
	}

}

func (h *Handler) writeError(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	code errCode,
	message string,
) {
	h.writeJSON(
		ctx,
		w,
		status,
		errorResponse{
			Code:    code,
			Message: message,
		},
	)
}
