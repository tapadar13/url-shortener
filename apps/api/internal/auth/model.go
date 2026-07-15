package auth

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailRequired       = errors.New("email is required")
	ErrEmailInvalid        = errors.New("email is invalid")
	ErrPasswordRequired    = errors.New("password is required")
	ErrPasswordTooShort    = errors.New("password must be at least 12 characters")
	ErrPasswordTooLong     = errors.New("password must not exceed 72 bytes")
	ErrPasswordHashInvalid = errors.New("password hash is invalid")
	ErrPasswordMismatch    = errors.New("password does not match")
)

const (
	minPasswordLength = 12
	maxEmailLength    = 320
	maxPasswordBytes  = 72
	bcryptCost        = bcrypt.DefaultCost
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NormalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return "", ErrEmailRequired
	}
	if len(normalized) > maxEmailLength {
		return "", ErrEmailInvalid
	}
	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized {
		return "", ErrEmailInvalid
	}
	return normalized, nil
}

func HashPassword(password string) (string, error) {
	if err := validatePassword(password); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePassword(password, encodedHash string) error {
	if err := validatePassword(password); err != nil {
		return err
	}
	if strings.TrimSpace(encodedHash) == "" {
		return ErrPasswordHashInvalid
	}
	if err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return ErrPasswordHashInvalid
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}
	if len([]rune(password)) < minPasswordLength {
		return ErrPasswordTooShort
	}
	if len([]byte(password)) > maxPasswordBytes {
		return ErrPasswordTooLong
	}
	return nil
}
