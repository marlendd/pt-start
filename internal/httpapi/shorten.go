package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/marlendd/pt-start/internal/shortener"
)

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var payload shortenRequest

	if err := decoder.Decode(&payload); err != nil {
		h.writeDecodeError(r, w, err)
		return
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.writeDecodeError(r, w, err)
		return
	}

	code, err := h.shortener.Shorten(r.Context(), payload.URL)
	if err != nil {
		if errors.Is(err, shortener.ErrInvalidURL) {
			h.writeError(
				r.Context(),
				w,
				http.StatusBadRequest,
				errCodeInvalidURL,
				"invalid URL",
			)
			return
		}

		h.logger.ErrorContext(
			r.Context(),
			"failed to shorten URL",
			"error",
			err,
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

	shortURL := h.baseURL + "/" + code

	w.Header().Set("Location", shortURL)

	h.writeJSON(
		r.Context(),
		w,
		http.StatusCreated,
		shortenResponse{
			ShortURL: shortURL,
		},
	)
}

func (h *Handler) writeDecodeError(
	r *http.Request,
	w http.ResponseWriter,
	err error,
) {
	var maxBytesError *http.MaxBytesError

	if errors.As(err, &maxBytesError) {
		h.writeError(
			r.Context(),
			w,
			http.StatusRequestEntityTooLarge,
			errCodePayloadTooLarge,
			"request body is too large",
		)
		return
	}

	h.writeError(
		r.Context(),
		w,
		http.StatusBadRequest,
		errCodeInvalidRequest,
		"invalid request body",
	)
}
