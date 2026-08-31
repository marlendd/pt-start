package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandlerHealth(t *testing.T) {
	handler := NewHandler(
		shortenerStub{},
		nil,
		newTestLogger(),
		"http://localhost:8080",
	)

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	handler.Health(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}

func TestHandlerReady(t *testing.T) {
	checker := readinessCheckerStub{
		pingFn: func(ctx context.Context) error {
			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			require.LessOrEqual(t, time.Until(deadline), readinessTimeout)

			return nil
		},
	}

	handler := NewHandler(
		shortenerStub{},
		checker,
		newTestLogger(),
		"http://localhost:8080",
	)

	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()

	handler.Ready(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}

func TestHandlerReadyReturnsNotReady(t *testing.T) {
	testCases := []struct {
		name      string
		readiness ReadinessChecker
	}{
		{
			name:      "checker is not configured",
			readiness: nil,
		},
		{
			name: "check fails",
			readiness: readinessCheckerStub{
				pingFn: func(context.Context) error {
					return errors.New("database unavailable")
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewHandler(
				shortenerStub{},
				tc.readiness,
				newTestLogger(),
				"http://localhost:8080",
			)

			request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			recorder := httptest.NewRecorder()

			handler.Ready(recorder, request)

			require.Equal(
				t,
				http.StatusServiceUnavailable,
				recorder.Code,
			)
			require.JSONEq(
				t,
				`{
					"code":"not_ready",
					"message":"service is not ready"
				}`,
				recorder.Body.String(),
			)
		})
	}
}
