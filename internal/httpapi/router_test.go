package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouterRoutesShorten(t *testing.T) {
	service := shortenerStub{
		shortenFn: func(
			context.Context,
			string,
		) (string, error) {
			return "abc12345", nil
		},
	}

	handler := NewHandler(
		service,
		nil,
		newTestLogger(),
		"http://localhost:8080",
	)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/shorten",
		strings.NewReader(
			`{"url":"https://example.com"}`,
		),
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code)
}

func TestRouterSetsCodePathValue(t *testing.T) {
	const code = "abc12345"

	service := shortenerStub{
		resolveFn: func(
			_ context.Context,
			receivedCode string,
		) (string, error) {
			require.Equal(t, code, receivedCode)

			return "https://example.com", nil
		},
	}

	handler := NewHandler(
		service,
		nil,
		newTestLogger(),
		"http://localhost:8080",
	)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/"+code,
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusFound, recorder.Code)
	require.Equal(
		t,
		"https://example.com",
		recorder.Header().Get("Location"),
	)
}

func TestRouterRejectsWrongMethod(t *testing.T) {
	handler := NewHandler(
		shortenerStub{},
		nil,
		newTestLogger(),
		"http://localhost:8080",
	)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/shorten",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(
		t,
		http.StatusMethodNotAllowed,
		recorder.Code,
	)
	require.Equal(
		t,
		http.MethodPost,
		recorder.Header().Get("Allow"),
	)
	require.JSONEq(
		t,
		`{
			"code": "method_not_allowed",
			"message": "method not allowed"
		}`,
		recorder.Body.String(),
	)
}

func TestRouterReturnsNotFound(t *testing.T) {
	handler := NewHandler(
		shortenerStub{},
		nil,
		newTestLogger(),
		"http://localhost:8080",
	)
	router := NewRouter(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/foo/bar",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.JSONEq(
		t,
		`{
			"code": "not_found",
			"message": "resource not found"
		}`,
		recorder.Body.String(),
	)
}

func TestRouterRoutesHealthChecks(t *testing.T) {
	checker := readinessCheckerStub{
		pingFn: func(context.Context) error {
			return nil
		},
	}

	handler := NewHandler(
		shortenerStub{},
		checker,
		newTestLogger(),
		"http://localhost:8080",
	)
	router := NewRouter(handler)

	for _, path := range []string{"/healthz", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusOK, recorder.Code)
			require.JSONEq(
				t,
				`{"status":"ok"}`,
				recorder.Body.String(),
			)
		})
	}
}
