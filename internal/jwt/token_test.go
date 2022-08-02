package jwt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func TestItSignsClaimsWithSuccess(t *testing.T) {
	key := "super-secret-key"
	p := NewProcessor(key)
	u := uuid.New()
	claims := &CustomClaims{
		PrincipalID: u.String(),
		Email:       "foo@bar.com",
	}

	s, err := p.Sign(context.Background(), claims)
	if err != nil {
		t.Fatalf("unexpected error signing claims, error %v", err)
	}

	if len(s) == 0 {
		t.Error("empty token string")
	}
}

func TestItValidatesJWTSignedClaims(t *testing.T) {
	key := "super-secret-key"
	p := NewProcessor(key)
	u := uuid.New()
	claims := &CustomClaims{
		PrincipalID: u.String(),
		Email:       "foo@bar.com",
	}

	s, _ := p.Sign(context.Background(), claims)

	if err := p.Validate(context.Background(), s); err != nil {
		t.Fatalf("unexpected error validating token, error %v", err)
	}
}

func TestItFailsOnExpiredTokenValidation(t *testing.T) {
	key := "super-secret-key"
	u := uuid.New()
	c := &CustomClaims{
		PrincipalID: u.String(),
		Email:       "foo@bar.com",
	}
	past := time.Now().Add(-time.Hour * 100)
	c.IssuedAt = past.Unix()
	c.ExpiresAt = past.Add(time.Hour * 24).Unix()
	c.Id = fmt.Sprintf("%s%d", c.PrincipalID, time.Now().UnixNano())
	c.Issuer = issuerName
	c.Subject = subject

	tkn, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(key))
	if err != nil {
		t.Fatalf("unable to create token, error %v", err)
	}

	p := NewProcessor(key)
	err = p.Validate(context.Background(), tkn)
	if err == nil {
		t.Fatal("expected validation error")
	}
}
