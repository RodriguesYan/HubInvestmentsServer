package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfig struct {
	ServerAddr string
	Timeout    time.Duration
}

func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		ServerAddr: "localhost:50051",
		Timeout:    30 * time.Second,
	}
}

// Common gRPC connection helper used by all clients
func dialGRPC(config *ClientConfig) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, config.ServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", config.ServerAddr, err)
	}

	return conn, nil
}

// ClientManager provides easy access to all gRPC clients
type ClientManager struct {
	Auth     *AuthClient
	Order    *OrderClient
	Position *PositionClient
}

func NewClientManager() *ClientManager {
	config := DefaultConfig()
	return NewClientManagerWithConfig(config)
}

func NewClientManagerWithConfig(config *ClientConfig) *ClientManager {
	return &ClientManager{
		Auth:     NewAuthClient(config),
		Order:    NewOrderClient(config),
		Position: NewPositionClient(config),
	}
}

func (cm *ClientManager) Close() error {
	var errs []error

	if err := cm.Auth.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := cm.Order.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := cm.Position.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}

	return nil
}
