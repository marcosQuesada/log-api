package proto

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestItSuccedsOnAuthorizationHeaderFound(t *testing.T) {
	v := &fakeRequestValidator{}
	a := NewJWTAuthAdapter(v)

	token := "fake_jwt_token"
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", token)

	if _, err := a.Interceptor(ctx, nil, nil, nopUnaryHandler); err != nil {
		t.Fatal("expected validation error")
	}
}

func TestItFailsOnAuthorizationHeaderNotFound(t *testing.T) {
	v := &fakeRequestValidator{}
	a := NewJWTAuthAdapter(v)

	_, err := a.Interceptor(context.Background(), nil, nil, nopUnaryHandler)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !errors.Is(err, ErrNoMetadataProvided) {
		t.Errorf("unexpected error type, got %v", err)
	}
}

type fakeRequestValidator struct {
	rawToken string
}

func (f *fakeRequestValidator) Validate(c context.Context, rawToken string) error {
	f.rawToken = rawToken
	return nil
}

func nopUnaryHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}
