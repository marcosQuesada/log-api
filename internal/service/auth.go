package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/marcosQuesada/log-api/internal/jwt"
	v1 "github.com/marcosQuesada/log-api/internal/proto/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type signer interface {
	Sign(ctx context.Context, j *jwt.CustomClaims) (string, error)
}

type User struct {
	ID string
}

type UserRepository interface {
	Get(ctx context.Context, user, password string) (*User, error)
}

type auth struct {
	signer     signer
	repository UserRepository

	v1.UnsafeAuthServiceServer
}

func NewAuth(s signer, u UserRepository) *auth {
	return &auth{
		signer:     s,
		repository: u,
	}
}

func (a *auth) Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error) {
	t, err := a.login(ctx, r.Username, r.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to login, error %v", err))
	}

	return &v1.LoginResponse{Token: t}, nil
}

func (a *auth) login(ctx context.Context, email, password string) (token string, err error) {
	u, err := a.repository.Get(ctx, email, password)
	if err != nil {
		return "", fmt.Errorf("unable to find email %s error %w", email, err)
	}

	tkn, err := a.signer.Sign(ctx, &jwt.CustomClaims{
		PrincipalID: u.ID,
		Email:       email,
	})

	if err != nil {
		return "", fmt.Errorf("unable to sign claims, error %w", err)
	}

	return tkn, nil
}

var errEmptyUserCredentials = errors.New("empty user credentials")

type userFakeRepository struct{}

func NewAuthFakeRepository() UserRepository {
	return &userFakeRepository{}
}

func (a *userFakeRepository) Get(ctx context.Context, user, password string) (*User, error) {
	if user == "" || password == "" {
		return nil, errEmptyUserCredentials
	}
	return &User{ID: uuid.New().String()}, nil
}
