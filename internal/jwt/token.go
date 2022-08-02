package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	issuerName          = "Log API"
	subject             = "Logger"
	jwtSigningAlgorithm = "HS256"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidJWTClaims = errors.New("invalid jwt claims expectation")
)

type CustomClaims struct {
	PrincipalID string `json:"principal_id"`
	Email       string `json:"email"`
	jwt.StandardClaims
}

type Processor struct {
	key string
}

// NewProcessor instantiates jwt signer
func NewProcessor(k string) *Processor {
	return &Processor{
		key: k,
	}
}

// Sign Claims with shared key
func (s *Processor) Sign(ctx context.Context, c *CustomClaims) (string, error) {
	c.IssuedAt = time.Now().Unix()
	c.ExpiresAt = time.Now().Add(time.Hour * 24).Unix()
	c.Id = fmt.Sprintf("%s%d", c.PrincipalID, time.Now().UnixNano())
	c.Issuer = issuerName
	c.Subject = subject

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(s.key))
}

// Validate checks token signature
func (s *Processor) Validate(c context.Context, rawToken string) error {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwtSigningAlgorithm {
			return nil, ErrInvalidToken
		}
		return []byte(s.key), nil
	}

	cl := &CustomClaims{}
	token, err := jwt.ParseWithClaims(rawToken, cl, keyFunc)
	if err != nil {
		return fmt.Errorf("unable to parse jwt token, error %w", err)
	}

	if token == nil || !token.Valid {
		return ErrInvalidToken
	}

	if cl == nil || cl.Issuer != issuerName {
		return ErrInvalidJWTClaims
	}

	return nil
}
