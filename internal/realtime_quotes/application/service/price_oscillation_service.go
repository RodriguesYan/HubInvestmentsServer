package service

import (
	"HubInvestments/internal/realtime_quotes/domain/model"
	"HubInvestments/internal/realtime_quotes/domain/service"
	"context"
	"log"
	"math/rand"
	"sync"
	"time"
)

type PriceOscillationService struct {
	assetDataService *service.AssetDataService
	subscribers      []chan map[string]*model.AssetQuote
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	ticker           *time.Ticker
}

func NewPriceOscillationService(assetDataService *service.AssetDataService) *PriceOscillationService {
	ctx, cancel := context.WithCancel(context.Background())

	return &PriceOscillationService{
		assetDataService: assetDataService,
		subscribers:      make([]chan map[string]*model.AssetQuote, 0),
		ctx:              ctx,
		cancel:           cancel,
		ticker:           time.NewTicker(2 * time.Second),
	}
}

func (s *PriceOscillationService) Start() {
	go s.oscillatePrices()
	log.Println("Price oscillation service started - prices will update every 2 seconds")
}

func (s *PriceOscillationService) Stop() {
	s.cancel()
	s.ticker.Stop()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, subscriber := range s.subscribers {
		close(subscriber)
	}
	s.subscribers = nil

	log.Println("Price oscillation service stopped")
}

func (s *PriceOscillationService) Subscribe() <-chan map[string]*model.AssetQuote {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscriber := make(chan map[string]*model.AssetQuote, 10)
	s.subscribers = append(s.subscribers, subscriber)

	return subscriber
}

func (s *PriceOscillationService) oscillatePrices() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.ticker.C:
			s.updatePrices()
		}
	}
}

func (s *PriceOscillationService) updatePrices() {
	assets := s.assetDataService.GetRandomAssets(5)

	for _, quote := range assets {
		newPrice := s.calculateNewPrice(quote)
		quote.UpdatePrice(newPrice)
	}

	s.notifySubscribers(assets)
}

// Simulate the price oscillation
func (s *PriceOscillationService) calculateNewPrice(quote *model.AssetQuote) float64 {
	// Generate random oscillation between -1% and +1%
	oscillationPercent := (rand.Float64() - 0.5) * 2 * 0.01 // -0.01 to +0.01 (±1%)

	// Apply oscillation to base price
	newPrice := quote.BasePrice * (1 + oscillationPercent)

	// Ensure price doesn't go below $1.00
	if newPrice < 1.00 {
		newPrice = 1.00
	}

	return newPrice
}

func (s *PriceOscillationService) notifySubscribers(assets map[string]*model.AssetQuote) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, subscriber := range s.subscribers {
		select {
		case subscriber <- assets:
		default:
			// Skip if subscriber channel is full to avoid blocking
		}
	}
}
