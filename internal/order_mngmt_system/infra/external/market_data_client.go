package external

import (
	"context"
	"fmt"
	"time"

	"github.com/RodriguesYan/hub-proto-contracts/monolith"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// IMarketDataClient defines the interface needed by order management domain services
// This interface follows dependency inversion principle for the domain layer
type IMarketDataClient interface {
	// GetAssetDetails retrieves detailed information about an asset
	GetAssetDetails(ctx context.Context, symbol string) (*AssetDetails, error)

	// ValidateSymbol checks if a symbol exists and is tradeable
	ValidateSymbol(ctx context.Context, symbol string) (bool, error)

	// GetCurrentPrice retrieves the current market price for a symbol
	GetCurrentPrice(ctx context.Context, symbol string) (float64, error)

	// IsMarketOpen checks if the market is currently open for trading
	IsMarketOpen(ctx context.Context, symbol string) (bool, error)

	// GetTradingHours retrieves trading hours information for a symbol
	// In future, we need to create a service only for handling trading hours
	GetTradingHours(ctx context.Context, symbol string) (*TradingHours, error)

	// GetBatchMarketData retrieves market data for multiple symbols
	GetBatchMarketData(ctx context.Context, symbols []string) ([]MarketDataResponse, error)

	// Close closes the underlying connections
	Close() error
}

// MarketDataResponse represents a simplified market data response
type MarketDataResponse struct {
	Symbol      string
	CompanyName string
	LastQuote   float64
	Category    string
}

// AssetDetails represents detailed information about a tradeable asset
type AssetDetails struct {
	Symbol       string
	Name         string
	Category     AssetCategory
	LastQuote    float64
	IsActive     bool
	IsTradeable  bool
	MaxOrderSize float64
	PriceStep    float64
	LastUpdated  time.Time
}

// AssetCategory represents the category of an asset
type AssetCategory int32

const (
	AssetCategoryStock AssetCategory = iota
	AssetCategoryBond
	AssetCategoryCrypto
	AssetCategoryFund
	AssetCategoryETF
)

// TradingHours represents trading hours information
type TradingHours struct {
	Symbol          string
	MarketOpen      time.Time
	MarketClose     time.Time
	IsOpen          bool
	NextOpenTime    time.Time
	NextCloseTime   time.Time
	Timezone        string
	ExtendedHours   bool
	PreMarketOpen   time.Time
	PostMarketClose time.Time
}

// MarketDataClient wraps the gRPC client and adapts it for order management needs
type MarketDataClient struct {
	conn       *grpc.ClientConn
	grpcClient monolith.MarketDataServiceClient
	timeout    time.Duration
}

// MarketDataClientConfig holds configuration for the market data client
type MarketDataClientConfig struct {
	ServerAddress string
	Timeout       time.Duration
}

// NewMarketDataClient creates a new market data client using the gRPC infrastructure
func NewMarketDataClient(config MarketDataClientConfig) (IMarketDataClient, error) {
	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.ServerAddress == "" {
		config.ServerAddress = "localhost:50054"
	}

	// Create gRPC connection
	conn, err := grpc.Dial(
		config.ServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(config.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to market data service: %w", err)
	}

	grpcClient := monolith.NewMarketDataServiceClient(conn)

	return &MarketDataClient{
		conn:       conn,
		grpcClient: grpcClient,
		timeout:    config.Timeout,
	}, nil
}

// NewMarketDataClientWithDefaults creates a client with default configuration
func NewMarketDataClientWithDefaults() (IMarketDataClient, error) {
	return NewMarketDataClient(MarketDataClientConfig{
		ServerAddress: "localhost:50054",
		Timeout:       30 * time.Second,
	})
}

// Close closes the gRPC connection
func (c *MarketDataClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetAssetDetails retrieves detailed information about an asset
func (c *MarketDataClient) GetAssetDetails(ctx context.Context, symbol string) (*AssetDetails, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Get market data via gRPC
	req := &monolith.GetMarketDataRequest{Symbol: symbol}
	resp, err := c.grpcClient.GetMarketData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data for symbol %s: %w", symbol, err)
	}

	if resp.MarketData == nil {
		return nil, fmt.Errorf("no data found for symbol %s", symbol)
	}

	// Convert to AssetDetails
	data := resp.MarketData
	assetDetails := &AssetDetails{
		Symbol:       data.Symbol,
		Name:         data.CompanyName,
		Category:     mapCategoryFromMarketData(int(data.Category)),
		LastQuote:    data.CurrentPrice,
		IsActive:     true, // Assume active if we got data
		IsTradeable:  c.isSymbolTradeable(data),
		MaxOrderSize: c.getMaxOrderSize(int(data.Category)),
		PriceStep:    c.getPriceStep(int(data.Category)),
		LastUpdated:  time.Now(),
	}

	return assetDetails, nil
}

// ValidateSymbol checks if a symbol exists and is tradeable
func (c *MarketDataClient) ValidateSymbol(ctx context.Context, symbol string) (bool, error) {
	assetDetails, err := c.GetAssetDetails(ctx, symbol)
	if err != nil {
		// If we can't get asset details, symbol is invalid
		return false, nil
	}

	// Symbol is valid if we found it and it's tradeable
	return assetDetails.IsTradeable, nil
}

// GetCurrentPrice retrieves the current market price for a symbol
func (c *MarketDataClient) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	assetDetails, err := c.GetAssetDetails(ctx, symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to get current price for symbol %s: %w", symbol, err)
	}

	return assetDetails.LastQuote, nil
}

// GetBatchMarketData retrieves market data for multiple symbols
func (c *MarketDataClient) GetBatchMarketData(ctx context.Context, symbols []string) ([]MarketDataResponse, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Get batch market data via gRPC
	req := &monolith.GetBatchMarketDataRequest{Symbols: symbols}
	resp, err := c.grpcClient.GetBatchMarketData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch market data: %w", err)
	}

	// Convert to MarketDataResponse
	result := make([]MarketDataResponse, 0, len(resp.MarketData))
	for _, data := range resp.MarketData {
		// Convert category int32 to string representation
		categoryStr := fmt.Sprintf("%d", data.Category)
		result = append(result, MarketDataResponse{
			Symbol:      data.Symbol,
			CompanyName: data.CompanyName,
			LastQuote:   data.CurrentPrice,
			Category:    categoryStr,
		})
	}

	return result, nil
}

// IsMarketOpen checks if the market is currently open for trading
func (c *MarketDataClient) IsMarketOpen(ctx context.Context, symbol string) (bool, error) {
	tradingHours, err := c.GetTradingHours(ctx, symbol)
	if err != nil {
		return false, fmt.Errorf("failed to check market hours for symbol %s: %w", symbol, err)
	}

	return tradingHours.IsOpen, nil
}

// GetTradingHours retrieves trading hours information for a symbol
func (c *MarketDataClient) GetTradingHours(ctx context.Context, symbol string) (*TradingHours, error) {
	assetDetails, err := c.GetAssetDetails(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset details for trading hours: %w", err)
	}

	// Calculate trading hours based on asset category and current time
	// This is a simplified implementation - in production you'd have more sophisticated logic
	now := time.Now()
	tradingHours := &TradingHours{
		Symbol:   symbol,
		Timezone: "America/New_York",
	}

	// Set trading hours based on asset category
	switch assetDetails.Category {
	case AssetCategoryStock, AssetCategoryETF:
		tradingHours.MarketOpen = c.getTodayTime(7, 30)      // 9:30 AM
		tradingHours.MarketClose = c.getTodayTime(22, 0)     // 4:00 PM
		tradingHours.PreMarketOpen = c.getTodayTime(4, 0)    // 4:00 AM
		tradingHours.PostMarketClose = c.getTodayTime(20, 0) // 8:00 PM
		tradingHours.ExtendedHours = true

	case AssetCategoryCrypto:
		// Crypto markets are open 24/7
		tradingHours.MarketOpen = c.getTodayTime(0, 0)
		tradingHours.MarketClose = c.getTodayTime(23, 59)
		tradingHours.IsOpen = true
		tradingHours.ExtendedHours = false
		return tradingHours, nil
	default:
		// Default to stock market hours
		tradingHours.MarketOpen = c.getTodayTime(7, 30)
		tradingHours.MarketClose = c.getTodayTime(22, 0)
		tradingHours.ExtendedHours = false
	}

	// Determine if market is currently open
	tradingHours.IsOpen = c.isMarketCurrentlyOpen(now, tradingHours)

	// Set next open/close times
	tradingHours.NextOpenTime = c.getNextOpenTime(now, tradingHours)
	tradingHours.NextCloseTime = c.getNextCloseTime(now, tradingHours)

	return tradingHours, nil
}

// Helper methods

func mapCategoryFromMarketData(category int) AssetCategory {
	// Map from market data category to order management category
	// This mapping should be adjusted based on your actual category values
	switch category {
	case 0:
		return AssetCategoryStock
	case 1:
		return AssetCategoryBond
	case 2:
		return AssetCategoryCrypto
	case 3:
		return AssetCategoryFund
	case 4:
		return AssetCategoryETF
	default:
		return AssetCategoryStock // Default fallback
	}
}

func (c *MarketDataClient) isSymbolTradeable(data *monolith.MarketData) bool {
	// Basic tradeability check - in production you'd have more sophisticated rules
	return data.CurrentPrice > 0 && data.Symbol != ""
}

func (c *MarketDataClient) getMaxOrderSize(category int) float64 {
	// Maximum order sizes based on asset category
	switch AssetCategory(category) {
	case AssetCategoryStock, AssetCategoryETF:
		return 1000000.0 // 1M shares max
	case AssetCategoryCrypto:
		return 1000000.0 // 1M units max
	case AssetCategoryBond:
		return 100000000.0 // $100M max for bonds
	default:
		return 1000000.0
	}
}

func (c *MarketDataClient) getPriceStep(category int) float64 {
	// Price increments based on asset category
	switch AssetCategory(category) {
	case AssetCategoryStock, AssetCategoryETF:
		return 0.01 // $0.01 increments
	case AssetCategoryCrypto:
		return 0.00000001 // Very small increments for crypto
	case AssetCategoryBond:
		return 0.125 // 1/8 point increments for bonds
	default:
		return 0.01
	}
}

func (c *MarketDataClient) getTodayTime(hour, minute int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
}

func (c *MarketDataClient) isMarketCurrentlyOpen(now time.Time, hours *TradingHours) bool {
	// Skip weekends for non-crypto assets
	// if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
	// 	return false
	// }

	// Check if current time is within trading hours
	return now.After(hours.MarketOpen) && now.Before(hours.MarketClose)
}

func (c *MarketDataClient) getNextOpenTime(now time.Time, hours *TradingHours) time.Time {
	if hours.IsOpen {
		// If market is open, next open is tomorrow
		return hours.MarketOpen.AddDate(0, 0, 1)
	}

	// If market is closed today but hasn't opened yet, return today's open
	if now.Before(hours.MarketOpen) {
		return hours.MarketOpen
	}

	// Market is closed for the day, return tomorrow's open
	return hours.MarketOpen.AddDate(0, 0, 1)
}

func (c *MarketDataClient) getNextCloseTime(now time.Time, hours *TradingHours) time.Time {
	if hours.IsOpen {
		// If market is open, next close is today
		return hours.MarketClose
	}

	// Market is closed, next close is when it opens and closes again
	nextOpen := c.getNextOpenTime(now, hours)
	return time.Date(nextOpen.Year(), nextOpen.Month(), nextOpen.Day(),
		hours.MarketClose.Hour(), hours.MarketClose.Minute(), 0, 0, nextOpen.Location())
}
