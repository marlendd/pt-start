package shortener

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

const (
	maxURLLength       = 2048
	maxShortenAttempts = 5
)

type Service struct {
	repository Repository
	generator  Generator
}

func NewService(repository Repository, generator Generator) *Service {
	return &Service{
		repository: repository,
		generator:  generator,
	}
}

func (s *Service) Shorten(
	ctx context.Context,
	originalURL string,
) (string, error) {
	if err := validateURL(originalURL); err != nil {
		return "", err
	}

	for range maxShortenAttempts {
		code, err := s.generator.Generate()
		if err != nil {
			return "", fmt.Errorf("generate short code: %w", err)
		}

		err = s.repository.Save(ctx, code, originalURL)
		if err == nil {
			return code, nil
		}

		if errors.Is(err, ErrCodeExists) {
			continue
		}

		return "", fmt.Errorf("save shortened URL: %w", err)
	}

	return "", fmt.Errorf(
		"generate unique short code after %d attempts: %w",
		maxShortenAttempts,
		ErrCodeExists,
	)
}

func (s *Service) Resolve(
	ctx context.Context,
	code string,
) (string, error) {
	if !isValidCode(code) {
		return "", ErrNotFound
	}

	originalURL, err := s.repository.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrNotFound
		}

		return "", fmt.Errorf("find original URL: %w", err)
	}

	return originalURL, nil
}

func validateURL(originalURL string) error {
	if originalURL == "" || len(originalURL) > maxURLLength {
		return ErrInvalidURL
	}

	if strings.IndexFunc(originalURL, unicode.IsSpace) >= 0 {
		return ErrInvalidURL
	}

	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return ErrInvalidURL
	}

	if parsedURL.Hostname() == "" {
		return ErrInvalidURL
	}

	if !strings.EqualFold(parsedURL.Scheme, "http") &&
		!strings.EqualFold(parsedURL.Scheme, "https") {
		return ErrInvalidURL
	}

	return nil
}

func isValidCode(code string) bool {
	if len(code) != codeLength {
		return false
	}

	for _, char := range code {
		if !strings.ContainsRune(base62Alphabet, char) {
			return false
		}
	}

	return true
}
