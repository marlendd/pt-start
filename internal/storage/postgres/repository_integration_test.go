package postgres_test

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marlendd/pt-start/internal/shortener"
	postgresstore "github.com/marlendd/pt-start/internal/storage/postgres"
	"github.com/stretchr/testify/require"
)

func newTestRepository(t *testing.T) *postgresstore.Repository {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(t.Context(), databaseURL)
	require.NoError(t, err)

	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(t.Context()))

	_, err = pool.Exec(t.Context(), "TRUNCATE TABLE links")
	require.NoError(t, err)

	return postgresstore.NewRepository(pool)
}

func TestRepositorySaveAndFindByCode(t *testing.T) {
	repository := newTestRepository(t)

	const (
		code        = "abc12345"
		originalURL = "https://example.com/page"
	)

	err := repository.Save(t.Context(), code, originalURL)
	require.NoError(t, err)

	resolvedURL, err := repository.FindByCode(
		t.Context(),
		code,
	)
	
	require.NoError(t, err)
	require.Equal(t, originalURL, resolvedURL)
}

func TestRepositorySaveReturnsCodeExists(t *testing.T) {
	repository := newTestRepository(t)

	const code = "abc12345"

	err := repository.Save(
		t.Context(),
		code,
		"https://example.com/first",
	)
	require.NoError(t, err)

	err = repository.Save(
		t.Context(),
		code,
		"https://example.com/second",
	)

	require.ErrorIs(t, err, shortener.ErrCodeExists)

	resolvedURL, err := repository.FindByCode(
		t.Context(),
		code,
	)

	require.NoError(t, err)
	require.Equal(t, "https://example.com/first", resolvedURL)
}

func TestRepositoryFindByCodeReturnsNotFound(t *testing.T) {
	repository := newTestRepository(t)

	resolvedURL, err := repository.FindByCode(
		t.Context(),
		"abc12345",
	)

	require.ErrorIs(t, err, shortener.ErrNotFound)
	require.Empty(t, resolvedURL)
}