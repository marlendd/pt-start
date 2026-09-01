package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRequestLogging(t *testing.T) {
	testCases := []struct {
		name           string
		writeStatus    bool
		status         int
		expectedStatus int
	}{
		{
			name:           "explicit status",
			writeStatus:    true,
			status:         http.StatusCreated,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "implicit OK status",
			writeStatus:    false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var logOutput bytes.Buffer

			logger := slog.New(
				slog.NewJSONHandler(
					&logOutput,
					nil,
				),
			)

			const responseBody = "response"

			next := http.HandlerFunc(
				func(w http.ResponseWriter, _ *http.Request) {
					if tc.writeStatus {
						w.WriteHeader(tc.status)
					}

					_, err := io.WriteString(
						w,
						responseBody,
					)
					require.NoError(t, err)
				},
			)

			handler := RequestLogging(logger)(next)

			request := httptest.NewRequest(
				http.MethodGet,
				"/test?foo=bar",
				nil,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			require.Equal(
				t,
				tc.expectedStatus,
				recorder.Code,
			)
			require.Equal(
				t,
				responseBody,
				recorder.Body.String(),
			)

			var record struct {
				Level         string        `json:"level"`
				Message       string        `json:"msg"`
				Method        string        `json:"method"`
				Path          string        `json:"path"`
				Status        int           `json:"status"`
				ResponseBytes int           `json:"response_bytes"`
				Duration      time.Duration `json:"duration"`
				RemoteAddress string        `json:"remote_address"`
			}

			require.NoError(
				t,
				json.Unmarshal(
					logOutput.Bytes(),
					&record,
				),
			)

			require.Equal(t, "INFO", record.Level)
			require.Equal(
				t,
				"HTTP request completed",
				record.Message,
			)
			require.Equal(t, http.MethodGet, record.Method)
			require.Equal(t, "/test", record.Path)
			require.Equal(
				t,
				tc.expectedStatus,
				record.Status,
			)
			require.Equal(
				t,
				len(responseBody),
				record.ResponseBytes,
			)
			require.Positive(t, record.Duration)
			require.Equal(
				t,
				request.RemoteAddr,
				record.RemoteAddress,
			)
		})
	}
}
