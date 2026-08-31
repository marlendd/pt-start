package shortener

import "errors"

var (
	ErrInvalidURL = errors.New("invalid URL")
	ErrNotFound   = errors.New("short URL not found")
	ErrCodeExists = errors.New("short code already exists")
)
