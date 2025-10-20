package grpc

import (
	"context"
	"fmt"

	authpb "github.com/RodriguesYan/hub-proto-contracts/auth"
	"google.golang.org/grpc"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
	config *ClientConfig
}

func NewAuthClient(config *ClientConfig) *AuthClient {
	return &AuthClient{
		config: config,
	}
}

func (c *AuthClient) connect() error {
	if c.conn != nil {
		return nil
	}

	conn, err := dialGRPC(c.config)
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = authpb.NewAuthServiceClient(conn)
	return nil
}

func (c *AuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *AuthClient) Login(email, password string) (*authpb.LoginResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	req := &authpb.LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := c.client.Login(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Auth.Login failed: %w", err)
	}

	return resp, nil
}

func (c *AuthClient) ValidateToken(token string) (*authpb.ValidateTokenResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	req := &authpb.ValidateTokenRequest{
		Token: token,
	}

	resp, err := c.client.ValidateToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Auth.ValidateToken failed: %w", err)
	}

	return resp, nil
}
