package client

import (
	"context"
	"fmt"
	"time"

	"HubInvestments/internal/market_data/domain/model"

	monolithpb "github.com/RodriguesYan/hub-proto-contracts/monolith"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// IMarketDataGRPCClient defines the interface for the gRPC client
type IMarketDataGRPCClient interface {
	GetMarketData(ctx context.Context, symbols []string) ([]model.MarketDataModel, error)
	Close() error
}

// MarketDataGRPCClient implements the gRPC client for market data service
type MarketDataGRPCClient struct {
	conn   *grpc.ClientConn
	client monolithpb.MarketDataServiceClient
}

// MarketDataGRPCClientConfig holds configuration for the gRPC client
type MarketDataGRPCClientConfig struct {
	ServerAddress string
	Timeout       time.Duration
}

// NewMarketDataGRPCClient creates a new gRPC client for market data service
func NewMarketDataGRPCClient(config MarketDataGRPCClientConfig) (IMarketDataGRPCClient, error) {
	// Set default timeout if not provided
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Set default server address if not provided
	if config.ServerAddress == "" {
		config.ServerAddress = "localhost:50054"
	}

	// Establish connection to gRPC server
	conn, err := grpc.Dial(config.ServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(config.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", config.ServerAddress, err)
	}

	// Create the gRPC client
	client := monolithpb.NewMarketDataServiceClient(conn)

	return &MarketDataGRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetMarketData retrieves market data for the given symbols via gRPC
func (c *MarketDataGRPCClient) GetMarketData(ctx context.Context, symbols []string) ([]model.MarketDataModel, error) {
	// Create request using GetBatchMarketData for multiple symbols
	req := &monolithpb.GetBatchMarketDataRequest{
		Symbols: symbols,
	}

	// Make gRPC call
	resp, err := c.client.GetBatchMarketData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data via gRPC: %w", err)
	}

	// Convert gRPC response to domain models
	result := make([]model.MarketDataModel, len(resp.MarketData))
	for i, data := range resp.MarketData {
		result[i] = model.MarketDataModel{
			Symbol:    data.Symbol,
			Name:      data.CompanyName,
			LastQuote: float32(data.CurrentPrice),
			Category:  int(data.Category),
		}
	}

	return result, nil
}

// Close closes the gRPC connection
func (c *MarketDataGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// NewMarketDataGRPCClientWithDefaults creates a client with default configuration
func NewMarketDataGRPCClientWithDefaults() (IMarketDataGRPCClient, error) {
	config := MarketDataGRPCClientConfig{
		ServerAddress: "localhost:50054",
		Timeout:       30 * time.Second,
	}
	return NewMarketDataGRPCClient(config)
}
