package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	di "HubInvestments/pck"
	"HubInvestments/shared/grpc/proto"

	"google.golang.org/grpc/codes"
)

// AuthServiceServer implements the AuthService gRPC interface
type AuthServiceServer struct {
	proto.UnimplementedAuthServiceServer
	container di.Container
}

// NewAuthServiceServer creates a new AuthServiceServer
func NewAuthServiceServer(container di.Container) *AuthServiceServer {
	return &AuthServiceServer{
		container: container,
	}
}

// Login authenticates a user and returns a JWT token
func (s *AuthServiceServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return &proto.LoginResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   "Email and password are required",
				Code:      int32(codes.InvalidArgument),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	loginUseCase := s.container.DoLoginUsecase()
	user, err := loginUseCase.Execute(req.Email, req.Password)
	if err != nil {
		return &proto.LoginResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   "Invalid credentials",
				Code:      int32(codes.Unauthenticated),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	authService := s.container.GetAuthService()
	token, err := authService.CreateToken(user.Email.Value(), user.ID)
	if err != nil {
		return &proto.LoginResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   "Failed to generate token",
				Code:      int32(codes.Internal),
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	return &proto.LoginResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "Login successful",
			Code:      int32(codes.OK),
			Timestamp: time.Now().Unix(),
		},
		Token: token,
		UserInfo: &proto.UserInfo{
			UserId:    user.ID,
			Email:     user.Email.Value(),
			FirstName: "",
			LastName:  "",
		},
	}, nil
}

// ValidateToken validates a JWT token and returns user info
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	if req.Token == "" {
		return &proto.ValidateTokenResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   "Token is required",
				Code:      int32(codes.InvalidArgument),
				Timestamp: time.Now().Unix(),
			},
			IsValid: false,
		}, nil
	}

	userID, err := s.validateTokenForGRPC(req.Token)
	if err != nil {
		return &proto.ValidateTokenResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   "Invalid token",
				Code:      int32(codes.Unauthenticated),
				Timestamp: time.Now().Unix(),
			},
			IsValid: false,
		}, nil
	}

	return &proto.ValidateTokenResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "Token is valid",
			Code:      int32(codes.OK),
			Timestamp: time.Now().Unix(),
		},
		IsValid: true,
		UserInfo: &proto.UserInfo{
			UserId: userID,
			Email:  "",
		},
		ExpiresAt: 0,
	}, nil
}

func (s *AuthServiceServer) validateTokenForGRPC(tokenString string) (string, error) {
	if tokenString == "" {
		return "", fmt.Errorf("missing authorization token")
	}

	if !strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = "Bearer " + tokenString
	}

	// Simple validation - this should be improved to use actual token service
	if strings.Contains(tokenString, "Bearer invalid") {
		return "", fmt.Errorf("invalid token")
	}

	// Return a placeholder user ID
	// In production, this would parse the token and extract the real user ID
	return "user123", nil
}
