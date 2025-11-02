package interceptors

import (
	"context"
	"strings"

	"HubInvestments/internal/auth"
	"HubInvestments/internal/auth/token"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// AuthInterceptor provides JWT authentication for gRPC requests
type AuthInterceptor struct {
	authService  auth.IAuthService
	tokenService token.ITokenService
}

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor() *AuthInterceptor {
	tokenService := token.NewTokenService()
	authService := auth.NewAuthService(tokenService)

	return &AuthInterceptor{
		authService:  authService,
		tokenService: tokenService,
	}
}

// verifyTokenForGRPC verifies JWT token for gRPC without HTTP ResponseWriter
func (interceptor *AuthInterceptor) verifyTokenForGRPC(tokenString string) (string, error) {
	if tokenString == "" {
		return "", status.Error(codes.Unauthenticated, "missing authorization token")
	}

	claims, err := interceptor.tokenService.ValidateToken(tokenString)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid token: "+err.Error())
	}

	userId, ok := claims["userId"].(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "invalid token claims")
	}

	return userId, nil
}

// UnaryInterceptor provides authentication for unary gRPC calls
func (interceptor *AuthInterceptor) UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	//TODO: Check what to do with it when migrating to microsservics
	// Allow internal service-to-service calls without authentication
	// Check if this is an internal call by checking the peer address
	if interceptor.isInternalCall(ctx) {
		// For internal calls, add a system user context and continue
		ctx = context.WithValue(ctx, "userId", "system")
		return handler(ctx, req)
	}

	// Extract metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Get authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := authHeaders[0]

	// Check for Bearer token format
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
	}

	// Extract token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	// Add Bearer prefix back for token service
	fullTokenString := "Bearer " + tokenString

	// Verify token using our custom method
	userId, err := interceptor.verifyTokenForGRPC(fullTokenString)
	if err != nil {
		return nil, err // Already a gRPC status error
	}

	// Add userId to context for use in handlers
	ctx = context.WithValue(ctx, "userId", userId)

	// Continue with the request
	return handler(ctx, req)
}

// StreamInterceptor provides authentication for streaming gRPC calls
func (interceptor *AuthInterceptor) StreamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// Extract metadata from stream context
	ctx := stream.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Get authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := authHeaders[0]

	// Check for Bearer token format
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return status.Error(codes.Unauthenticated, "invalid authorization header format")
	}

	// Extract token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return status.Error(codes.Unauthenticated, "missing token")
	}

	// Add Bearer prefix back for token service
	fullTokenString := "Bearer " + tokenString

	// Verify token
	userId, err := interceptor.verifyTokenForGRPC(fullTokenString)
	if err != nil {
		return err // Already a gRPC status error
	}

	// Add userId to context
	ctx = context.WithValue(ctx, "userId", userId)

	// Create new stream with updated context
	wrappedStream := &wrappedServerStream{
		ServerStream: stream,
		ctx:          ctx,
	}

	// Continue with the request
	return handler(srv, wrappedStream)
}

// wrappedServerStream wraps grpc.ServerStream to override context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// isInternalCall checks if the gRPC call is coming from an internal service
func (interceptor *AuthInterceptor) isInternalCall(ctx context.Context) bool {
	// Check if the call is coming from localhost (internal service)
	if p, ok := peer.FromContext(ctx); ok {
		addr := p.Addr.String()
		// Allow calls from localhost or loopback addresses
		return strings.Contains(addr, "127.0.0.1") || strings.Contains(addr, "::1") || strings.Contains(addr, "localhost")
	}
	return false
}
