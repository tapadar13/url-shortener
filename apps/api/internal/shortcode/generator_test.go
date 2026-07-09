package shortcode

import (
	"errors"
	"io"
	"testing"
)

func TestGeneratorGenerateReturnsValidCode(t *testing.T) {
	t.Parallel()

	generator := NewGenerator()

	code, err := generator.Generate(7)
	if err != nil {
		t.Fatalf("expected code to be generated: %v", err)
	}

	if len(code) != 7 {
		t.Fatalf("expected code length 7, got %d", len(code))
	}

	if err := Validate(code); err != nil {
		t.Fatalf("expected generated code to be valid: %v", err)
	}
}

func TestGeneratorRejectsInvalidLength(t *testing.T) {
	t.Parallel()

	generator := NewGenerator()

	if _, err := generator.Generate(MinLength - 1); !errors.Is(err, ErrTooShort) {
		t.Fatalf("expected too short error, got %v", err)
	}

	if _, err := generator.Generate(MaxLength + 1); !errors.Is(err, ErrTooLong) {
		t.Fatalf("expected too long error, got %v", err)
	}
}

func TestGeneratorUsesProvidedRandomReader(t *testing.T) {
	t.Parallel()

	generator := NewGeneratorWithReader(zeroReader{})

	code, err := generator.Generate(MinLength)
	if err != nil {
		t.Fatalf("expected code to be generated: %v", err)
	}

	if code != "0000" {
		t.Fatalf("expected deterministic code, got %q", code)
	}
}

func TestGeneratorReturnsReaderError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("reader failed")
	generator := NewGeneratorWithReader(failingReader{err: expectedErr})

	_, err := generator.Generate(MinLength)
	if !errors.Is(err, ErrGenerateFailed) {
		t.Fatalf("expected generate failure, got %v", err)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped reader error, got %v", err)
	}
}

func TestNewGeneratorWithReaderFallsBackWhenReaderIsNil(t *testing.T) {
	t.Parallel()

	generator := NewGeneratorWithReader(nil)

	code, err := generator.Generate(MinLength)
	if err != nil {
		t.Fatalf("expected code to be generated: %v", err)
	}

	if err := Validate(code); err != nil {
		t.Fatalf("expected generated code to be valid: %v", err)
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}

	return len(p), nil
}

type failingReader struct {
	err error
}

func (r failingReader) Read(_ []byte) (int, error) {
	return 0, r.err
}

var _ io.Reader = zeroReader{}
var _ io.Reader = failingReader{}
