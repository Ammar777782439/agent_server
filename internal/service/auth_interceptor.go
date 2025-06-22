package service

import (
	"context"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

const (
	AuthTokenEnvVar = "AGENT_SERVER_AUTH_TOKEN"
	AuthHeaderKey   = "authorization"
)

// AuthInterceptor is a gRPC unary interceptor for token-based auth.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Allow unauthenticated access to reflection service
		if info.FullMethod == "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		tokens := md[AuthHeaderKey]
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing auth token")
		}

		expectedToken := os.Getenv(AuthTokenEnvVar)
		if expectedToken == "" {
			return nil, status.Error(codes.Internal, "server auth token not configured")
		}

		if tokens[0] != expectedToken {
			return nil, status.Error(codes.Unauthenticated, "invalid auth token")
		}
		return handler(ctx, req)
	}
}
