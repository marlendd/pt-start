package httpapi

import (
	"errors"
	"net/http"

	"github.com/marlendd/pt-start/internal/shortener"
)

func (h *Handler) Redirect(
	w http.ResponseWriter,
	r *http.Request,
) {
	code := r.PathValue("code")

	originalURL, err := h.shortener.Resolve(
		r.Context(),
		code,
	)
	if err != nil {
		if errors.Is(err, shortener.ErrNotFound) {
			h.writeError(
				r.Context(),
				w,
				http.StatusNotFound,
				errCodeNotFound,
				"short URL not found",
			)
			return
		}

		h.logger.ErrorContext(
			r.Context(),
			"failed to resolve short URL",
			"code", code,
			"error", err,
		)

		h.writeError(
			r.Context(),
			w,
			http.StatusInternalServerError,
			errCodeInternal,
			"internal server error",
		)
		return
	}

	http.Redirect(
		w,
		r,
		originalURL,
		http.StatusFound,
	)
}
