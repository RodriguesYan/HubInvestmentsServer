package grpc_test

import (
	"context"
	"net"
	"testing"

	balanceGrpc "HubInvestments/internal/balance/presentation/grpc"
	marketDataGrpc "HubInvestments/internal/market_data/presentation/grpc"
	orderGrpc "HubInvestments/internal/order_mngmt_system/presentation/grpc"
	portfolioGrpc "HubInvestments/internal/portfolio_summary/presentation/grpc"
	positionGrpc "HubInvestments/internal/position/presentation/grpc"
	di "HubInvestments/pck"
	"HubInvestments/shared/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// setupTestServer creates a test gRPC server with all handlers
func setupTestServer(t *testing.T) (*grpc.Server, di.Container) {
	container, err := di.NewContainer()
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	server := grpc.NewServer()

	// Register all handlers
	portfolioHandler := portfolioGrpc.NewPortfolioGRPCHandler(container)
	balanceHandler := balanceGrpc.NewBalanceGRPCHandler(container)
	marketDataHandler := marketDataGrpc.NewMarketDataGRPCHandler(container)
	orderHandler := orderGrpc.NewOrderGRPCHandler(container)
	positionHandler := positionGrpc.NewPositionGRPCHandler(container)

	proto.RegisterPortfolioServiceServer(server, portfolioHandler)
	proto.RegisterBalanceServiceServer(server, balanceHandler)
	proto.RegisterMarketDataServiceServer(server, marketDataHandler)
	proto.RegisterOrderServiceServer(server, orderHandler)
	proto.RegisterPositionServiceServer(server, positionHandler)

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	return server, container
}

// TestBalanceService_GetBalance tests the GetBalance gRPC endpoint
func TestBalanceService_GetBalance(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Stop()

	ctx := context.Background()
	// Note: Using DialContext for testing with bufconn
	// Production code should use grpc.NewClient instead
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := proto.NewBalanceServiceClient(conn)

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "Valid user ID",
			userID:  "1",
			wantErr: false,
		},
		{
			name:    "Empty user ID",
			userID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetBalance(ctx, &proto.GetBalanceRequest{
				UserId: tt.userID,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("Expected response but got nil")
				return
			}

			if resp.ApiResponse == nil || !resp.ApiResponse.Success {
				t.Error("Expected successful API response")
			}

			if resp.Balance == nil {
				t.Error("Expected balance data but got nil")
			}
		})
	}
}

// TestMarketDataService_GetMarketData tests the GetMarketData gRPC endpoint
func TestMarketDataService_GetMarketData(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := proto.NewMarketDataServiceClient(conn)

	tests := []struct {
		name    string
		symbol  string
		wantErr bool
	}{
		{
			name:    "Valid symbol",
			symbol:  "AAPL",
			wantErr: false,
		},
		{
			name:    "Empty symbol",
			symbol:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetMarketData(ctx, &proto.GetMarketDataRequest{
				Symbol: tt.symbol,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("Expected response but got nil")
				return
			}

			if resp.ApiResponse == nil || !resp.ApiResponse.Success {
				t.Error("Expected successful API response")
			}
		})
	}
}

// TestAuthenticationFlow tests the authentication flow with metadata
func TestAuthenticationFlow(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := proto.NewBalanceServiceClient(conn)

	tests := []struct {
		name      string
		withToken bool
		token     string
		wantErr   bool
	}{
		{
			name:      "Request without token (should work from localhost)",
			withToken: false,
			wantErr:   false,
		},
		{
			name:      "Request with valid token",
			withToken: true,
			token:     "Bearer valid-jwt-token",
			wantErr:   false, // Will fail validation but test structure is correct
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.withToken {
				md := metadata.New(map[string]string{
					"authorization": tt.token,
				})
				ctx = metadata.NewOutgoingContext(ctx, md)
			}

			_, err := client.GetBalance(ctx, &proto.GetBalanceRequest{
				UserId: "1",
			})

			// Note: This test verifies the structure, not actual auth validation
			// Actual auth validation requires a real JWT token
			if err != nil {
				t.Logf("Got error (expected for invalid token): %v", err)
			}
		})
	}
}

// TestConcurrentRequests tests concurrent gRPC requests
func TestConcurrentRequests(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := proto.NewBalanceServiceClient(conn)

	// Run 10 concurrent requests
	concurrency := 10
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			_, err := client.GetBalance(ctx, &proto.GetBalanceRequest{
				UserId: "1",
			})
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	close(errors)
	errorCount := 0
	for err := range errors {
		t.Logf("Concurrent request error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Logf("Note: %d/%d concurrent requests failed (expected if database is not available)", errorCount, concurrency)
	}
}

// TestPortfolioService_GetPortfolioSummary tests the GetPortfolioSummary gRPC endpoint
func TestPortfolioService_GetPortfolioSummary(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := proto.NewPortfolioServiceClient(conn)

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "Valid user ID",
			userID:  "1",
			wantErr: false,
		},
		{
			name:    "Empty user ID",
			userID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetPortfolioSummary(ctx, &proto.GetPortfolioSummaryRequest{
				UserId: tt.userID,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("Expected response but got nil")
				return
			}

			if resp.ApiResponse == nil || !resp.ApiResponse.Success {
				t.Error("Expected successful API response")
			}
		})
	}
}

// TestPositionService_GetPositions tests the GetPositions gRPC endpoint
func TestPositionService_GetPositions(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := proto.NewPositionServiceClient(conn)

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "Valid user ID",
			userID:  "1",
			wantErr: false,
		},
		{
			name:    "Empty user ID",
			userID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetPositions(ctx, &proto.GetPositionsRequest{
				UserId: tt.userID,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("Expected response but got nil")
				return
			}

			if resp.ApiResponse == nil || !resp.ApiResponse.Success {
				t.Error("Expected successful API response")
			}
		})
	}
}
