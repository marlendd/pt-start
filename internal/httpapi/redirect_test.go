package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/marlendd/pt-start/internal/shortener"
)

func TestHandlerRedirectSuccess(t *testing.T) {
	const (
		code        = "abc12345"
		originalURL = "https://example.com/page"
	)

	service := shortenerStub{
		resolveFn: func(
			_ context.Context,
			receivedCode string,
		) (string, error) {
			require.Equal(t, code, receivedCode)

			return originalURL, nil
		},
	}

	handler := NewHandler(
		service,
		newTestLogger(),
		"http://localhost:8080",
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/"+code,
		nil,
	)
	request.SetPathValue("code", code)

	recorder := httptest.NewRecorder()

	handler.Redirect(recorder, request)

	require.Equal(t, http.StatusFound, recorder.Code)
	require.Equal(
		t,
		originalURL,
		recorder.Header().Get("Location"),
	)
}

func TestHandlerRedirectHandlesErrors(t *testing.T) {
	testCases := []struct {
		name           string
		serviceError   error
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:           "not found",
			serviceError:   shortener.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedJSON: `{
				"code": "not_found",
				"message": "short URL not found"
			}`,
		},
		{
			name:           "internal error",
			serviceError:   errors.New("database unavailable"),
			expectedStatus: http.StatusInternalServerError,
			expectedJSON: `{
				"code": "internal_error",
				"message": "internal server error"
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := shortenerStub{
				resolveFn: func(
					context.Context,
					string,
				) (string, error) {
					return "", tc.serviceError
				},
			}

			handler := NewHandler(
				service,
				newTestLogger(),
				"http://localhost:8080",
			)

			request := httptest.NewRequest(
				http.MethodGet,
				"/abc12345",
				nil,
			)
			request.SetPathValue("code", "abc12345")

			recorder := httptest.NewRecorder()

			handler.Redirect(recorder, request)

			require.Equal(t, tc.expectedStatus, recorder.Code)
			require.JSONEq(
				t,
				tc.expectedJSON,
				recorder.Body.String(),
			)
			require.Empty(
				t,
				recorder.Header().Get("Location"),
			)
		})
	}
}
