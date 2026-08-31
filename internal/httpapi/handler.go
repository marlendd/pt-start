package httpapi

import (
	"context"
	"log/slog"
	"strings"
)

const maxRequestBodySize = 4 << 10 // 4 KiB

type errCode string

const (
	errCodeInvalidRequest  errCode = "invalid_request"
	errCodeInvalidURL      errCode = "invalid_url"
	errCodePayloadTooLarge errCode = "payload_too_large"
	errCodeNotFound        errCode = "not_found"
	errCodeInternal        errCode = "internal_error"
)

type Shortener interface {
	Shorten(
		ctx context.Context,
		originalURL string,
	) (string, error)

	Resolve(
		ctx context.Context,
		code string,
	) (string, error)
}

type Handler struct {
	shortener Shortener
	logger    *slog.Logger
	baseURL   string
}

func NewHandler(
	shortener Shortener,
	logger *slog.Logger,
	baseURL string,
) *Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return &Handler{
		shortener: shortener,
		logger:    logger,
		baseURL:   strings.TrimRight(baseURL, "/"),
	}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

type errorResponse struct {
	Code    errCode `json:"code"`
	Message string  `json:"message"`
}
