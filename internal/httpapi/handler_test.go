package httpapi

import (
	"context"
	"io"
	"log/slog"
)

type shortenerStub struct {
	shortenFn func(context.Context, string) (string, error)
	resolveFn func(context.Context, string) (string, error)
}

func (s shortenerStub) Shorten(
	ctx context.Context,
	originalURL string,
) (string, error) {
	return s.shortenFn(ctx, originalURL)
}

func (s shortenerStub) Resolve(
	ctx context.Context,
	code string,
) (string, error) {
	return s.resolveFn(ctx, code)
}

func newTestLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)
}
