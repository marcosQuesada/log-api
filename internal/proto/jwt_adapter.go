package proto

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var ErrNoAuthorizationHeader = errors.New("no authorization header")
var ErrNoMetadataProvided = errors.New("metadata is not provided")

const (
	authHeader                = "authorization"
	bearerCleanOut            = "Bearer "
	unrestrictedLoginEndpoint = "/v1.AuthService/Login"
)

type requestValidator interface {
	Validate(c context.Context, rawToken string) error
}

type JWTAuthAdapter struct {
	validator requestValidator
}

func NewJWTAuthAdapter(v requestValidator) *JWTAuthAdapter {
	return &JWTAuthAdapter{validator: v}
}

// Interceptor defines a unary Interceptor that validates JWT tokens
func (a *JWTAuthAdapter) Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if info.FullMethod == unrestrictedLoginEndpoint {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrNoMetadataProvided
	}

	if len(md[authHeader]) == 0 {
		return nil, ErrNoAuthorizationHeader
	}

	tkn := md[authHeader][0]
	tkn = strings.ReplaceAll(tkn, bearerCleanOut, "")

	if err := a.validator.Validate(ctx, tkn); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization token")
	}

	return handler(ctx, req)
}
