package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	originalValue, existed := os.LookupEnv(key)

	require.NoError(t, os.Unsetenv(key))

	t.Cleanup(func() {
		if existed {
			require.NoError(
				t,
				os.Setenv(key, originalValue),
			)
			return
		}

		require.NoError(t, os.Unsetenv(key))
	})
}

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv(
		"DATABASE_URL",
		"postgres://localhost/test",
	)

	unsetEnv(t, "HTTP_ADDR")
	unsetEnv(t, "BASE_URL")
	unsetEnv(t, "SHUTDOWN_TIMEOUT")
	unsetEnv(t, "LOG_LEVEL")

	config, err := Load()

	require.NoError(t, err)
	require.Equal(t, ":8080", config.HTTPAddr)
	require.Equal(
		t,
		"http://localhost:8080",
		config.BaseURL,
	)
	require.Equal(
		t,
		"postgres://localhost/test",
		config.DatabaseURL,
	)
	require.Equal(
		t,
		10*time.Second,
		config.ShutdownTimeout,
	)
	require.Equal(t, "INFO", config.LogLevel)
}

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("HTTP_ADDR", "127.0.0.1:9090")
	t.Setenv("BASE_URL", "https://short.example/api///")
	t.Setenv(
		"DATABASE_URL",
		"postgres://localhost/custom",
	)
	t.Setenv("SHUTDOWN_TIMEOUT", "30s")
	t.Setenv("LOG_LEVEL", "DEBUG")

	config, err := Load()

	require.NoError(t, err)
	require.Equal(
		t,
		"127.0.0.1:9090",
		config.HTTPAddr,
	)
	require.Equal(
		t,
		"https://short.example/api",
		config.BaseURL,
	)
	require.Equal(
		t,
		"postgres://localhost/custom",
		config.DatabaseURL,
	)
	require.Equal(
		t,
		30*time.Second,
		config.ShutdownTimeout,
	)
	require.Equal(t, "DEBUG", config.LogLevel)
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("BASE_URL", "http://localhost:8080")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("LOG_LEVEL", "INFO")

	config, err := Load()

	require.Error(t, err)
	require.Empty(t, config.DatabaseURL)
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "invalid base URL",
			key:   "BASE_URL",
			value: "://invalid",
		},
		{
			name:  "base URL without host",
			key:   "BASE_URL",
			value: "http:///path",
		},
		{
			name:  "unsupported base URL scheme",
			key:   "BASE_URL",
			value: "ftp://example.com",
		},
		{
			name:  "base URL with query",
			key:   "BASE_URL",
			value: "https://example.com?foo=bar",
		},
		{
			name:  "base URL with fragment",
			key:   "BASE_URL",
			value: "https://example.com#section",
		},
		{
			name:  "invalid shutdown timeout",
			key:   "SHUTDOWN_TIMEOUT",
			value: "invalid",
		},
		{
			name:  "zero shutdown timeout",
			key:   "SHUTDOWN_TIMEOUT",
			value: "0s",
		},
		{
			name:  "negative shutdown timeout",
			key:   "SHUTDOWN_TIMEOUT",
			value: "-1s",
		},
		{
			name:  "invalid log level",
			key:   "LOG_LEVEL",
			value: "VERBOSE",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("HTTP_ADDR", ":8080")
			t.Setenv(
				"BASE_URL",
				"http://localhost:8080",
			)
			t.Setenv(
				"DATABASE_URL",
				"postgres://localhost/test",
			)
			t.Setenv("SHUTDOWN_TIMEOUT", "10s")
			t.Setenv("LOG_LEVEL", "INFO")

			t.Setenv(testCase.key, testCase.value)

			config, err := Load()

			require.Error(t, err)
			require.Empty(t, config)
		})
	}
}
