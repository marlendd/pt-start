package shortener

import "context"

type Repository interface {
	Save(
		ctx context.Context,
		code string,
		originalURL string,
	) error

	FindByCode(
		ctx context.Context,
		code string,
	) (string, error)
}
