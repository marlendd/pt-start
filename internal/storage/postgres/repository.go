package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/marlendd/pt-start/internal/shortener"
)

const uniqueViolationSQLState = "23505"

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) Save(
	ctx context.Context,
	code string,
	originalURL string,
) error {
	const query = `
		INSERT INTO links (code, original_url)
		VALUES ($1, $2)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		code,
		originalURL,
	)

	if err == nil {
		return nil
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) &&
		postgresError.Code == uniqueViolationSQLState {
		return shortener.ErrCodeExists
	}

	return fmt.Errorf("insert link: %w", err)
}

func (r *Repository) FindByCode(
	ctx context.Context,
	code string,
) (string, error) {
	const query = `
		SELECT original_url
		FROM links
		WHERE code = $1
	`

	var originalURL string

	err := r.pool.QueryRow(
		ctx,
		query,
		code,
	).Scan(&originalURL)
	if err == nil {
		return originalURL, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return "", shortener.ErrNotFound
	}

	return "", fmt.Errorf("find link by code: %w", err)
}

var _ shortener.Repository = (*Repository)(nil)
