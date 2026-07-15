package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenSecretRequired = errors.New("token secret is required")
	ErrTokenSecretWeak     = errors.New("token secret must be at least 32 characters")
	ErrTokenTTLInvalid     = errors.New("token TTL must be greater than zero")
	ErrTokenInvalid        = errors.New("token is invalid")
)

type TokenService struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time
}

type TokenOptions struct {
	Secret   string
	Issuer   string
	Audience string
	TTL      time.Duration
}

type TokenClaims struct {
	UserID string
}

func NewTokenService(options TokenOptions) (*TokenService, error) {
	if strings.TrimSpace(options.Secret) == "" {
		return nil, ErrTokenSecretRequired
	}
	if len(options.Secret) < 32 {
		return nil, ErrTokenSecretWeak
	}
	if options.TTL <= 0 {
		return nil, ErrTokenTTLInvalid
	}
	return &TokenService{
		secret:   []byte(options.Secret),
		issuer:   options.Issuer,
		audience: options.Audience,
		ttl:      options.TTL,
		now:      time.Now,
	}, nil
}

func newTokenService(options TokenOptions, now func() time.Time) (*TokenService, error) {
	service, err := NewTokenService(options)
	if err != nil {
		return nil, err
	}
	if now != nil {
		service.now = now
	}
	return service, nil
}

func (s *TokenService) Issue(userID string) (string, time.Time, error) {
	if s == nil || len(s.secret) == 0 {
		return "", time.Time{}, ErrTokenSecretRequired
	}
	if strings.TrimSpace(userID) == "" {
		return "", time.Time{}, errors.New("user ID is required")
	}
	now := s.now().UTC()
	expiresAt := now.Add(s.ttl)
	claims := jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Subject:   userID,
		Audience:  jwt.ClaimStrings{s.audience},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	encoded, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return encoded, expiresAt, nil
}

func (s *TokenService) Verify(encoded string) (TokenClaims, error) {
	if s == nil || len(s.secret) == 0 {
		return TokenClaims{}, ErrTokenSecretRequired
	}
	parsed, err := jwt.ParseWithClaims(encoded, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrTokenInvalid
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer), jwt.WithAudience(s.audience), jwt.WithLeeway(time.Second), jwt.WithTimeFunc(s.now))
	if err != nil || parsed == nil || !parsed.Valid {
		return TokenClaims{}, ErrTokenInvalid
	}
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || strings.TrimSpace(claims.Subject) == "" {
		return TokenClaims{}, ErrTokenInvalid
	}
	return TokenClaims{UserID: claims.Subject}, nil
}
