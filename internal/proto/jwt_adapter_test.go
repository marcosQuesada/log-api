package proto

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestItSucceedsOnAuthorizationHeaderFound(t *testing.T) {
	v := &fakeRequestValidator{}
	a := NewJWTAuthAdapter(v)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{"authorization": []string{"fake_jwt_token"}})
	if _, err := a.Interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/v1.FakeService/Foo"}, nopUnaryHandler); err != nil {
		t.Fatalf("unexpected validation error %v", err)
	}
}

func TestItFailsOnAuthorizationHeaderNotFound(t *testing.T) {
	v := &fakeRequestValidator{}
	a := NewJWTAuthAdapter(v)
	_, err := a.Interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/v1.FakeService/Foo"}, nopUnaryHandler)
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
