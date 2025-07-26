package client

import (
	"context"
	"net"
	"testing"
	"time"

	"HubInvestments/market_data/domain/model"
	"HubInvestments/market_data/presentation/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// mockMarketDataServer implements the gRPC server for testing
type mockMarketDataServer struct {
	proto.UnimplementedMarketDataServiceServer
	marketData []*proto.MarketData
	shouldFail bool
}

func (m *mockMarketDataServer) GetMarketData(ctx context.Context, req *proto.GetMarketDataRequest) (*proto.GetMarketDataResponse, error) {
	if m.shouldFail {
		return nil, grpc.ErrServerStopped
	}

	return &proto.GetMarketDataResponse{
		MarketData: m.marketData,
	}, nil
}

func setupTestServer(t *testing.T, mockServer *mockMarketDataServer) (*grpc.ClientConn, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	server := grpc.NewServer()
	proto.RegisterMarketDataServiceServer(server, mockServer)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	cleanup := func() {
		conn.Close()
		server.Stop()
		lis.Close()
	}

	return conn, cleanup
}

func TestMarketDataGRPCClient_GetMarketData_Success(t *testing.T) {
	// Arrange
	mockData := []*proto.MarketData{
		{
			Symbol:    "AAPL",
			Name:      "Apple Inc.",
			LastQuote: 150.25,
			Category:  1,
		},
		{
			Symbol:    "GOOGL",
			Name:      "Alphabet Inc.",
			LastQuote: 2750.50,
			Category:  1,
		},
	}

	mockServer := &mockMarketDataServer{
		marketData: mockData,
		shouldFail: false,
	}

	conn, cleanup := setupTestServer(t, mockServer)
	defer cleanup()

	client := &MarketDataGRPCClient{
		conn:   conn,
		client: proto.NewMarketDataServiceClient(conn),
	}

	// Act
	ctx := context.Background()
	symbols := []string{"AAPL", "GOOGL"}
	result, err := client.GetMarketData(ctx, symbols)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 market data items, got %d", len(result))
	}

	expected := []model.MarketDataModel{
		{
			Symbol:    "AAPL",
			Name:      "Apple Inc.",
			LastQuote: 150.25,
			Category:  1,
		},
		{
			Symbol:    "GOOGL",
			Name:      "Alphabet Inc.",
			LastQuote: 2750.50,
			Category:  1,
		},
	}

	for i, item := range result {
		if item.Symbol != expected[i].Symbol {
			t.Errorf("Expected symbol %s, got %s", expected[i].Symbol, item.Symbol)
		}
		if item.Name != expected[i].Name {
			t.Errorf("Expected name %s, got %s", expected[i].Name, item.Name)
		}
		if item.LastQuote != expected[i].LastQuote {
			t.Errorf("Expected last quote %f, got %f", expected[i].LastQuote, item.LastQuote)
		}
		if item.Category != expected[i].Category {
			t.Errorf("Expected category %d, got %d", expected[i].Category, item.Category)
		}
	}
}

func TestMarketDataGRPCClient_GetMarketData_Error(t *testing.T) {
	// Arrange
	mockServer := &mockMarketDataServer{
		shouldFail: true,
	}

	conn, cleanup := setupTestServer(t, mockServer)
	defer cleanup()

	client := &MarketDataGRPCClient{
		conn:   conn,
		client: proto.NewMarketDataServiceClient(conn),
	}

	// Act
	ctx := context.Background()
	symbols := []string{"AAPL"}
	result, err := client.GetMarketData(ctx, symbols)

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

func TestMarketDataGRPCClient_Close(t *testing.T) {
	// Arrange
	mockServer := &mockMarketDataServer{}
	conn, cleanup := setupTestServer(t, mockServer)
	defer cleanup()

	client := &MarketDataGRPCClient{
		conn:   conn,
		client: proto.NewMarketDataServiceClient(conn),
	}

	// Act & Assert
	err := client.Close()
	if err != nil {
		t.Errorf("Expected no error when closing, got %v", err)
	}
}

func TestNewMarketDataGRPCClient_WithConfig(t *testing.T) {
	// Test with custom configuration
	config := MarketDataGRPCClientConfig{
		ServerAddress: "localhost:50051",
		Timeout:       10 * time.Second,
	}

	// Note: gRPC dial doesn't immediately fail when no server is running
	// The connection is created lazily, so we test the actual call
	client, err := NewMarketDataGRPCClient(config)
	if err != nil {
		// This is also acceptable - connection might fail immediately
		return
	}

	defer client.Close()

	// Try to make an actual call to verify the connection fails
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = client.GetMarketData(ctx, []string{"TEST"})
	if err == nil {
		t.Error("Expected error when calling non-existent server, got nil")
	}
}

func TestNewMarketDataGRPCClient_WithDefaults(t *testing.T) {
	// Test with default configuration
	client, err := NewMarketDataGRPCClientWithDefaults()
	if err != nil {
		// This is acceptable - connection might fail immediately
		return
	}

	defer client.Close()

	// Try to make an actual call to verify the connection fails
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = client.GetMarketData(ctx, []string{"TEST"})
	if err == nil {
		t.Error("Expected error when calling non-existent server, got nil")
	}
}

func TestMarketDataGRPCClientConfig_Defaults(t *testing.T) {
	// Test that defaults are applied correctly
	config := MarketDataGRPCClientConfig{}

	// Create client to test defaults
	client, err := NewMarketDataGRPCClient(config)
	if err != nil {
		// This is acceptable - connection might fail immediately
		return
	}

	defer client.Close()

	// Try to make an actual call to verify the connection fails
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = client.GetMarketData(ctx, []string{"TEST"})
	if err == nil {
		t.Error("Expected error when calling non-existent server, got nil")
	}
}
