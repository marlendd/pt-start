package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPAddr        string        `env:"HTTP_ADDR" env-default:":8080"`
	BaseURL         string        `env:"BASE_URL" env-default:"http://localhost:8080"`
	DatabaseURL     string        `env:"DATABASE_URL" env-required:"true"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"10s"`
	LogLevel        string        `env:"LOG_LEVEL" env-default:"INFO"`
}

func Load() (Config, error) {
	var config Config

	if err := cleanenv.ReadEnv(&config); err != nil {
		return Config{}, fmt.Errorf(
			"read environment: %w",
			err,
		)
	}

	config.HTTPAddr = strings.TrimSpace(config.HTTPAddr)
	config.BaseURL = strings.TrimSpace(config.BaseURL)
	config.DatabaseURL = strings.TrimSpace(config.DatabaseURL)
	config.LogLevel = strings.TrimSpace(config.LogLevel)

	if err := validate(&config); err != nil {
		return Config{}, err
	}

	config.BaseURL = strings.TrimRight(
		config.BaseURL,
		"/",
	)

	return config, nil
}

func validate(config *Config) error {
	if config.DatabaseURL == "" {
		return fmt.Errorf(
			"DATABASE_URL must not be empty",
		)
	}

	if config.ShutdownTimeout <= 0 {
		return fmt.Errorf(
			"SHUTDOWN_TIMEOUT must be positive",
		)
	}

	parsedURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return fmt.Errorf("parse BASE_URL: %w", err)
	}

	if parsedURL.Hostname() == "" {
		return fmt.Errorf(
			"BASE_URL must contain a host",
		)
	}

	if parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return fmt.Errorf(
			"BASE_URL must not contain query or fragment",
		)
	}

	if !strings.EqualFold(parsedURL.Scheme, "http") &&
		!strings.EqualFold(parsedURL.Scheme, "https") {
		return fmt.Errorf(
			"BASE_URL must use http or https",
		)
	}

	var level slog.Level
	if err := level.UnmarshalText(
		[]byte(config.LogLevel),
	); err != nil {
		return fmt.Errorf("parse LOG_LEVEL: %w", err)
	}

	return nil
}
