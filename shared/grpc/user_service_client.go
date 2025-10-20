package grpc

import (
	"context"
	"fmt"
	"log"

	authpb "github.com/RodriguesYan/hub-proto-contracts/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserServiceClient wraps the gRPC client for User Management Service
type UserServiceClient struct {
	client authpb.UserServiceClient
	conn   *grpc.ClientConn
}

// NewUserServiceClient creates a new User Service gRPC client
func NewUserServiceClient(serviceAddress string) (*UserServiceClient, error) {
	if serviceAddress == "" {
		serviceAddress = "localhost:50052" // Default User Service gRPC port
	}

	// Create connection without deprecated WithBlock option
	conn, err := grpc.NewClient(
		serviceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create User Service client: %w", err)
	}

	// Initiate connection (non-blocking)
	conn.Connect()
	state := conn.GetState()

	log.Printf("✅ User Service client created for %s (state: %v)", serviceAddress, state)

	return &UserServiceClient{
		client: authpb.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

// ValidateToken validates a JWT token via User Service
func (c *UserServiceClient) ValidateToken(ctx context.Context, token string) (bool, string, string, error) {
	if token == "" {
		return false, "", "", fmt.Errorf("token cannot be empty")
	}

	resp, err := c.client.UserValidateToken(ctx, &authpb.UserValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return false, "", "", fmt.Errorf("failed to validate token: %w", err)
	}

	if !resp.Valid {
		return false, "", "", fmt.Errorf("invalid token: %s", resp.ErrorMessage)
	}

	return resp.Valid, resp.UserId, resp.Email, nil
}

// Login authenticates a user via User Service
func (c *UserServiceClient) Login(ctx context.Context, email, password string) (string, string, string, error) {
	if email == "" || password == "" {
		return "", "", "", fmt.Errorf("email and password are required")
	}

	resp, err := c.client.UserLogin(ctx, &authpb.UserLoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to login: %w", err)
	}

	if !resp.Success {
		return "", "", "", fmt.Errorf("login failed: %s", resp.ErrorMessage)
	}

	return resp.Token, resp.UserId, resp.Email, nil
}

// RegisterUser creates a new user via User Service
func (c *UserServiceClient) RegisterUser(ctx context.Context, email, password, firstName, lastName string) (string, error) {
	if email == "" || password == "" || firstName == "" || lastName == "" {
		return "", fmt.Errorf("all fields are required")
	}

	resp, err := c.client.RegisterUser(ctx, &authpb.RegisterUserRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("registration failed: %s", resp.ErrorMessage)
	}

	return resp.UserId, nil
}

// GetUserProfile retrieves user profile via User Service
func (c *UserServiceClient) GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	resp, err := c.client.GetUserProfile(ctx, &authpb.GetUserProfileRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to get profile: %s", resp.ErrorMessage)
	}

	return &UserProfile{
		UserID:        resp.UserId,
		Email:         resp.Email,
		FirstName:     resp.FirstName,
		LastName:      resp.LastName,
		IsActive:      resp.IsActive,
		EmailVerified: resp.EmailVerified,
	}, nil
}

// HealthCheck checks if User Service is healthy
func (c *UserServiceClient) HealthCheck(ctx context.Context) (bool, string, error) {
	resp, err := c.client.HealthCheck(ctx, &authpb.HealthCheckRequest{})
	if err != nil {
		return false, "", fmt.Errorf("health check failed: %w", err)
	}

	return resp.Healthy, resp.Version, nil
}

// Close closes the gRPC connection
func (c *UserServiceClient) Close() error {
	if c.conn != nil {
		log.Println("Closing User Service client connection...")
		return c.conn.Close()
	}
	return nil
}

// UserProfile represents a user profile returned from User Service
type UserProfile struct {
	UserID        string
	Email         string
	FirstName     string
	LastName      string
	IsActive      bool
	EmailVerified bool
}
