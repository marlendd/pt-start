package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/marlendd/pt-start/internal/shortener"
	"github.com/stretchr/testify/require"
)

func TestHandlerShortenSuccess(t *testing.T) {
	const (
		originalURL = "https://example.com/page"
		code        = "abc12345"
		baseURL     = "http://localhost:8080/"
		shortURL    = "http://localhost:8080/abc12345"
	)

	service := shortenerStub{
		shortenFn: func(
			_ context.Context,
			receivedURL string,
		) (string, error) {
			require.Equal(t, originalURL, receivedURL)

			return code, nil
		},
	}

	handler := NewHandler(
		service,
		newTestLogger(),
		baseURL,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		strings.NewReader(
			`{"url":"https://example.com/page"}`,
		),
	)

	recorder := httptest.NewRecorder()

	handler.Shorten(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, shortURL, recorder.Header().Get("Location"))
	require.Equal(
		t,
		"application/json; charset=utf-8",
		recorder.Header().Get("Content-Type"),
	)
	require.JSONEq(
		t,
		`{"short_url":"http://localhost:8080/abc12345"}`,
		recorder.Body.String(),
	)
}

func TestHandlerShortenRejectsInvalidBody(t *testing.T) {
	testCases := []struct {
		name           string
		body           string
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:           "empty body",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedJSON: `{
				"code": "invalid_request",
				"message": "invalid request body"
			}`,
		},
		{
			name:           "invalid JSON",
			body:           `{"url":`,
			expectedStatus: http.StatusBadRequest,
			expectedJSON: `{
				"code": "invalid_request",
				"message": "invalid request body"
			}`,
		},
		{
			name: "unknown field",
			body: `{
				"url": "https://example.com",
				"extra": true
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedJSON: `{
				"code": "invalid_request",
				"message": "invalid request body"
			}`,
		},
		{
			name:           "multiple JSON values",
			body:           `{"url":"https://example.com"} {}`,
			expectedStatus: http.StatusBadRequest,
			expectedJSON: `{
				"code": "invalid_request",
				"message": "invalid request body"
			}`,
		},
		{
			name: "body too large",
			body: `{"url":"` +
				strings.Repeat("a", maxRequestBodySize) +
				`"}`,
			expectedStatus: http.StatusRequestEntityTooLarge,
			expectedJSON: `{
				"code": "payload_too_large",
				"message": "request body is too large"
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serviceCalls := 0

			service := shortenerStub{
				shortenFn: func(
					ctx context.Context,
					s string,
				) (string, error) {
					serviceCalls++
					return "abc12345", nil
				},
			}

			handler := NewHandler(
				service,
				newTestLogger(),
				"http://localhost:8080",
			)

			request := httptest.NewRequest(
				http.MethodPost,
				"/shorten",
				strings.NewReader(tc.body),
			)

			recorder := httptest.NewRecorder()

			handler.Shorten(recorder, request)

			require.Equal(
				t,
				tc.expectedStatus,
				recorder.Code,
			)
			require.JSONEq(
				t,
				tc.expectedJSON,
				recorder.Body.String(),
			)
			require.Zero(t, serviceCalls)
		})
	}
}

func TestHandlerShortenHandlesServiceErrors(t *testing.T) {
	testCases := []struct {
		name           string
		serviceError   error
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:           "invalid URL",
			serviceError:   shortener.ErrInvalidURL,
			expectedStatus: http.StatusBadRequest,
			expectedJSON: `{
				"code": "invalid_url",
				"message": "invalid URL"
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
				shortenFn: func(
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
				http.MethodPost,
				"/shorten",
				strings.NewReader(
					`{"url":"https://example.com"}`,
				),
			)

			recorder := httptest.NewRecorder()

			handler.Shorten(recorder, request)

			require.Equal(
				t,
				tc.expectedStatus,
				recorder.Code,
			)
			require.JSONEq(
				t,
				tc.expectedJSON,
				recorder.Body.String(),
			)
			require.Empty(t, recorder.Header().Get("Location"))
		})
	}
}
