package shortcode

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var ErrGenerateFailed = errors.New("generate short code")

type Generator struct {
	reader io.Reader
}

func NewGenerator() Generator {
	return Generator{
		reader: rand.Reader,
	}
}

func NewGeneratorWithReader(reader io.Reader) Generator {
	if reader == nil {
		reader = rand.Reader
	}

	return Generator{
		reader: reader,
	}
}

func (g Generator) Generate(length int) (string, error) {
	if length < MinLength {
		return "", ErrTooShort
	}

	if length > MaxLength {
		return "", ErrTooLong
	}

	max := big.NewInt(int64(len(alphabet)))

	for {
		code := make([]byte, length)
		for i := range code {
			index, err := rand.Int(g.reader, max)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrGenerateFailed, err)
			}

			code[i] = alphabet[index.Int64()]
		}

		value := string(code)
		if IsReserved(value) {
			continue
		}

		return value, nil
	}
}
