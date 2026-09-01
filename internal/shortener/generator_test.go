package shortener

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomGeneratorGenerate(t *testing.T) {
	generator := RandomGenerator{}

	for range 100 {
		code, err := generator.Generate()

		require.NoError(t, err)
		require.Len(t, code, codeLength)

		for _, char := range code {
			require.Contains(t, base62Alphabet, string(char))
		}
	}
}
