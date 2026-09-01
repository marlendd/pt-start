package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecoveryHandlesPanic(t *testing.T) {
	var logOutput bytes.Buffer

	logger := slog.New(
		slog.NewJSONHandler(
			&logOutput,
			nil,
		),
	)

	next := http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {
			panic("test panic")
		},
	)

	handler := Recovery(logger)(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/panic",
		nil,
	)
	recorder := httptest.NewRecorder()

	require.NotPanics(t, func() {
		handler.ServeHTTP(recorder, request)
	})

	require.Equal(
		t,
		http.StatusInternalServerError,
		recorder.Code,
	)
	require.Equal(
		t,
		"application/json; charset=utf-8",
		recorder.Header().Get("Content-Type"),
	)
	require.JSONEq(
		t,
		`{
			"code": "internal_error",
			"message": "internal server error"
		}`,
		recorder.Body.String(),
	)

	var record struct {
		Level   string `json:"level"`
		Message string `json:"msg"`
		Panic   string `json:"panic"`
		Stack   string `json:"stack"`
	}

	require.NoError(
		t,
		json.Unmarshal(
			logOutput.Bytes(),
			&record,
		),
	)

	require.Equal(t, "ERROR", record.Level)
	require.Equal(t, "panic recovered", record.Message)
	require.Equal(t, "test panic", record.Panic)
	require.NotEmpty(t, record.Stack)
}

func TestRecoveryPassesSuccessfulResponse(t *testing.T) {
	var logOutput bytes.Buffer

	logger := slog.New(
		slog.NewJSONHandler(
			&logOutput,
			nil,
		),
	)

	next := http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	)

	handler := Recovery(logger)(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/success",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Empty(t, recorder.Body.String())
	require.Empty(t, logOutput.String())
}
