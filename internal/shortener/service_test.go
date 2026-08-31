package shortener

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	saveFn       func(context.Context, string, string) error
	findByCodeFn func(context.Context, string) (string, error)
}

func (s repositoryStub) Save(
	ctx context.Context,
	code string,
	originalURL string,
) error {
	return s.saveFn(ctx, code, originalURL)
}

func (s repositoryStub) FindByCode(
	ctx context.Context,
	code string,
) (string, error) {
	return s.findByCodeFn(ctx, code)
}

type generatorStub struct {
	generateFn func() (string, error)
}

func (s generatorStub) Generate() (string, error) {
	return s.generateFn()
}

func TestServiceShortenSuccess(t *testing.T) {
	const (
		originalURL  = "https://example.com/page"
		expectedCode = "abc12345"
	)

	repository := repositoryStub{
		saveFn: func(
			ctx context.Context,
			code string,
			savedURL string,
		) error {
			require.Equal(t, expectedCode, code)
			require.Equal(t, originalURL, savedURL)

			return nil
		},
	}

	generator := generatorStub{
		generateFn: func() (string, error) {
			return expectedCode, nil
		},
	}

	service := NewService(repository, generator)

	code, err := service.Shorten(t.Context(), originalURL)

	require.NoError(t, err)
	require.Equal(t, expectedCode, code)
}

func TestServiceShortenRejectsInvalidURL(t *testing.T) {
	testCases := map[string]string{
		"empty":              "",
		"without scheme":     "example.com",
		"unsupported scheme": "ftp://example.com",
		"without host":       "https:///some/path",
		"too long": "https://example.com/" +
			strings.Repeat("a", maxURLLength),
	}

	for name, originalURL := range testCases {
		t.Run(name, func(t *testing.T) {
			generatorCalls := 0
			saveCalls := 0

			generator := generatorStub{
				generateFn: func() (string, error) {
					generatorCalls++
					return "abc12345", nil
				},
			}

			repository := repositoryStub{
				saveFn: func(
					context.Context,
					string,
					string,
				) error {
					saveCalls++
					return nil
				},
			}

			service := NewService(repository, generator)

			code, err := service.Shorten(t.Context(), originalURL)

			require.ErrorIs(t, err, ErrInvalidURL)
			require.Empty(t, code)
			require.Zero(t, generatorCalls)
			require.Zero(t, saveCalls)
		})
	}
}

func TestServiceShortenRetriesAfterCollision(t *testing.T) {
	codes := []string{"first123", "second45"}
	generateCalls := 0
	saveCalls := 0

	generator := generatorStub{
		generateFn: func() (string, error) {
			code := codes[generateCalls]
			generateCalls++

			return code, nil
		},
	}

	repository := repositoryStub{
		saveFn: func(
			_ context.Context,
			code string,
			_ string,
		) error {
			require.Equal(t, codes[saveCalls], code)

			saveCalls++
			if saveCalls == 1 {
				return ErrCodeExists
			}

			return nil
		},
	}

	service := NewService(repository, generator)

	code, err := service.Shorten(
		t.Context(),
		"https://example.com",
	)

	require.NoError(t, err)
	require.Equal(t, "second45", code)
	require.Equal(t, 2, generateCalls)
	require.Equal(t, 2, saveCalls)
}

func TestServiceShortenReturnsGeneratorError(t *testing.T) {
	generatorErr := errors.New("random source unavailable")
	saveCalls := 0

	generator := generatorStub{
		generateFn: func() (string, error) {
			return "", generatorErr
		},
	}

	repository := repositoryStub{
		saveFn: func(
			context.Context,
			string,
			string,
		) error {
			saveCalls++
			return nil
		},
	}

	service := NewService(repository, generator)

	code, err := service.Shorten(
		t.Context(),
		"https://example.com",
	)

	require.ErrorIs(t, err, generatorErr)
	require.Empty(t, code)
	require.Zero(t, saveCalls)
}

func TestServiceShortenReturnsRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database unavailable")

	generator := generatorStub{
		generateFn: func() (string, error) {
			return "abc12345", nil
		},
	}

	repository := repositoryStub{
		saveFn: func(
			context.Context,
			string,
			string,
		) error {
			return repositoryErr
		},
	}

	service := NewService(repository, generator)

	code, err := service.Shorten(
		t.Context(),
		"https://example.com",
	)

	require.ErrorIs(t, err, repositoryErr)
	require.Empty(t, code)
}

func TestServiceShortenStopsAfterTooManyCollisions(t *testing.T) {
	generateCalls := 0
	saveCalls := 0

	generator := generatorStub{
		generateFn: func() (string, error) {
			generateCalls++
			return "abc12345", nil
		},
	}

	repository := repositoryStub{
		saveFn: func(
			context.Context,
			string,
			string,
		) error {
			saveCalls++
			return ErrCodeExists
		},
	}

	service := NewService(repository, generator)

	code, err := service.Shorten(
		t.Context(),
		"https://example.com",
	)

	require.ErrorIs(t, err, ErrCodeExists)
	require.Empty(t, code)
	require.Equal(t, maxShortenAttempts, generateCalls)
	require.Equal(t, maxShortenAttempts, saveCalls)
}

func TestServiceResolveSuccess(t *testing.T) {
	const (
		code        = "abc12345"
		originalURL = "https://example.com/page"
	)

	repository := repositoryStub{
		findByCodeFn: func(
			_ context.Context,
			receivedCode string,
		) (string, error) {
			require.Equal(t, code, receivedCode)

			return originalURL, nil
		},
	}

	service := NewService(repository, nil)

	resolvedURL, err := service.Resolve(t.Context(), code)

	require.NoError(t, err)
	require.Equal(t, originalURL, resolvedURL)
}

func TestServiceResolveReturnsNotFound(t *testing.T) {
	repository := repositoryStub{
		findByCodeFn: func(
			context.Context,
			string,
		) (string, error) {
			return "", ErrNotFound
		},
	}

	service := NewService(repository, nil)

	resolvedURL, err := service.Resolve(
		t.Context(),
		"abc12345",
	)

	require.ErrorIs(t, err, ErrNotFound)
	require.Empty(t, resolvedURL)
}

func TestServiceResolveReturnsRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database unavailable")

	repository := repositoryStub{
		findByCodeFn: func(
			context.Context,
			string,
		) (string, error) {
			return "", repositoryErr
		},
	}

	service := NewService(repository, nil)

	resolvedURL, err := service.Resolve(
		t.Context(),
		"abc12345",
	)

	require.ErrorIs(t, err, repositoryErr)
	require.Empty(t, resolvedURL)
}

func TestServiceResolveRejectsInvalidCode(t *testing.T) {
	testCases := map[string]string{
		"empty":              "",
		"too short":          "abc1234",
		"too long":           "abc123456",
		"unsupported symbol": "abc12-45",
	}

	for name, code := range testCases {
		t.Run(name, func(t *testing.T) {
			findCalls := 0

			repository := repositoryStub{
				findByCodeFn: func(
					context.Context,
					string,
				) (string, error) {
					findCalls++
					return "", nil
				},
			}

			service := NewService(repository, nil)

			resolvedURL, err := service.Resolve(
				t.Context(),
				code,
			)

			require.ErrorIs(t, err, ErrNotFound)
			require.Empty(t, resolvedURL)
			require.Zero(t, findCalls)
		})
	}
}
