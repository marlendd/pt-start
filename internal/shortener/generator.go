package shortener

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	codeLength     = 8
	base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Generator interface {
	Generate() (string, error)
}

type RandomGenerator struct{}

func (RandomGenerator) Generate() (string, error) {
	code := make([]byte, codeLength)
	alphabetSize := big.NewInt(int64(len(base62Alphabet)))

	for i := range code {
		index, err := rand.Int(rand.Reader, alphabetSize)
		if err != nil {
			return "", fmt.Errorf("generate random character: %w", err)
		}

		code[i] = base62Alphabet[index.Int64()]
	}

	return string(code), nil
}
