package usecase

import (
	marketDataClient "HubInvestments/internal/market_data/presentation/grpc/client"
	domain "HubInvestments/internal/position/domain/model"
	repository "HubInvestments/internal/position/domain/repository"
	service "HubInvestments/internal/position/domain/service"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type GetPositionAggregationUseCase struct {
	repo               repository.PositionRepository
	aggregationService service.PositionAggregationService
	marketDataClient   marketDataClient.IMarketDataGRPCClient
}

func NewGetPositionAggregationUseCase(repo repository.PositionRepository) *GetPositionAggregationUseCase {
	// Create market data client for fetching current prices
	mdClient, err := marketDataClient.NewMarketDataGRPCClient(marketDataClient.MarketDataGRPCClientConfig{
		ServerAddress: "localhost:50054",
		Timeout:       0, // Use default
	})
	if err != nil {
		log.Printf("Warning: Failed to create market data client: %v. Positions will show 0 for current prices.", err)
		mdClient = nil
	}

	return &GetPositionAggregationUseCase{
		repo:               repo,
		aggregationService: service.NewPositionAggregationService(),
		marketDataClient:   mdClient,
	}
}

// NewGetPositionAggregationUseCaseWithService allows dependency injection of the aggregation service for testing
func NewGetPositionAggregationUseCaseWithService(repo repository.PositionRepository, aggregationService service.PositionAggregationService) *GetPositionAggregationUseCase {
	return &GetPositionAggregationUseCase{
		repo:               repo,
		aggregationService: aggregationService,
	}
}

func (uc *GetPositionAggregationUseCase) Execute(userId string) (domain.AucAggregationModel, error) {
	userUUID, err := parseUserIDToUUID(userId)
	if err != nil {
		return domain.AucAggregationModel{}, fmt.Errorf("invalid user ID format '%s': %w", userId, err)
	}

	positions, err := uc.repo.FindByUserID(context.Background(), userUUID)
	if err != nil {
		return domain.AucAggregationModel{}, err
	}

	// Fetch current market prices for all symbols
	priceMap := uc.fetchMarketPrices(positions)

	// Convert positions to AssetModel for existing aggregation service
	assets := make([]domain.AssetModel, len(positions))
	for i, position := range positions {
		// Use current market price if available, otherwise fall back to stored CurrentPrice
		currentPrice := position.CurrentPrice
		if marketPrice, exists := priceMap[position.Symbol]; exists && marketPrice > 0 {
			currentPrice = marketPrice
		}

		assets[i] = domain.AssetModel{
			Symbol:       position.Symbol,
			Quantity:     float32(position.Quantity),
			AveragePrice: float32(position.AveragePrice),
			LastPrice:    float32(currentPrice),
			Category:     1,
		}
	}

	positionAggregations := uc.aggregationService.AggregateAssetsByCategory(assets)
	totalInvested, currentTotal := uc.aggregationService.CalculateTotals(assets)

	return domain.AucAggregationModel{
		TotalInvested:       totalInvested,
		CurrentTotal:        currentTotal,
		PositionAggregation: positionAggregations,
	}, nil
}

// fetchMarketPrices fetches current market prices for all position symbols
func (uc *GetPositionAggregationUseCase) fetchMarketPrices(positions []*domain.Position) map[string]float64 {
	if uc.marketDataClient == nil || len(positions) == 0 {
		return make(map[string]float64)
	}

	// Collect unique symbols
	symbolSet := make(map[string]bool)
	for _, pos := range positions {
		symbolSet[pos.Symbol] = true
	}

	symbols := make([]string, 0, len(symbolSet))
	for symbol := range symbolSet {
		symbols = append(symbols, symbol)
	}

	// Fetch market data for all symbols
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	marketDataList, err := uc.marketDataClient.GetMarketData(ctx, symbols)
	if err != nil {
		log.Printf("Warning: Failed to fetch market data for positions: %v", err)
		return make(map[string]float64)
	}

	// Build price map
	priceMap := make(map[string]float64)
	for _, md := range marketDataList {
		priceMap[md.Symbol] = float64(md.LastQuote)
	}

	return priceMap
}

// converts user ID string to UUID with flexible parsing
// Supports both UUID format strings and integer strings (for backward compatibility)
// MUST use the same format as command/helpers.go to ensure consistency!
func parseUserIDToUUID(userId string) (uuid.UUID, error) {
	// First, try parsing as a direct UUID
	if userUUID, err := uuid.Parse(userId); err == nil {
		return userUUID, nil
	}

	// If UUID parsing fails, try treating it as an integer and convert to UUID format
	// Uses the same format as command/helpers.go: 00000000-0000-0000-0000-000000000001
	if userInt, err := strconv.Atoi(userId); err == nil {
		// Convert integer to UUID format: 00000000-0000-0000-0000-000000000001
		uuidStr := fmt.Sprintf("00000000-0000-0000-0000-%012d", userInt)
		return uuid.Parse(uuidStr)
	}

	// If both parsing attempts fail, return error
	return uuid.Nil, fmt.Errorf("user ID '%s' cannot be parsed as UUID or integer", userId)
}
