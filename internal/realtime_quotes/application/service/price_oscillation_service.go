package service

import (
	"HubInvestments/internal/realtime_quotes/domain/model"
	"HubInvestments/internal/realtime_quotes/domain/service"
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	mathRand "math/rand"
	"sync"
	"time"
)

type Subscriber struct {
	channel chan map[string]*model.AssetQuote
	symbols map[string]bool
	id      string
}

type PriceOscillationService struct {
	assetDataService *service.AssetDataService
	subscribers      map[string]*Subscriber
	activeSymbols    map[string]int // symbol -> count of subscribers
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	ticker           *time.Ticker
}

func NewPriceOscillationService(assetDataService *service.AssetDataService) *PriceOscillationService {
	ctx, cancel := context.WithCancel(context.Background())

	return &PriceOscillationService{
		assetDataService: assetDataService,
		subscribers:      make(map[string]*Subscriber),
		activeSymbols:    make(map[string]int),
		ctx:              ctx,
		cancel:           cancel,
		ticker:           time.NewTicker(4 * time.Second), //Update quote every 4 seconds
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
		close(subscriber.channel)
	}
	s.subscribers = make(map[string]*Subscriber)
	s.activeSymbols = make(map[string]int)

	log.Println("Price oscillation service stopped")
}

func (s *PriceOscillationService) SubscribeToSymbols(symbols map[string]bool) (string, <-chan map[string]*model.AssetQuote) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscriberID := s.generateSubscriberID()

	subscriber := &Subscriber{
		channel: make(chan map[string]*model.AssetQuote, 100),
		symbols: make(map[string]bool),
		id:      subscriberID,
	}

	for symbol := range symbols {
		subscriber.symbols[symbol] = true
		s.activeSymbols[symbol]++
	}

	s.subscribers[subscriberID] = subscriber

	log.Printf("New subscriber %s for symbols: %v. Active symbols: %v",
		subscriberID, s.mapToSlice(symbols), s.getActiveSymbolsList())

	return subscriberID, subscriber.channel
}

func (s *PriceOscillationService) Unsubscribe(subscriberID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscriber, exists := s.subscribers[subscriberID]
	if !exists {
		return
	}

	for symbol := range subscriber.symbols {
		s.activeSymbols[symbol]--
		if s.activeSymbols[symbol] <= 0 {
			delete(s.activeSymbols, symbol)
		}
	}

	close(subscriber.channel)
	delete(s.subscribers, subscriberID)

	log.Printf("Unsubscribed %s. Active symbols: %v",
		subscriberID, s.getActiveSymbolsList())
}

// GetAllQuotes returns all available asset quotes for initial connection setup
func (s *PriceOscillationService) GetAllQuotes() map[string]*model.AssetQuote {
	// Get all available assets from the asset data service
	return s.assetDataService.GetAllAssets()
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
	s.mu.RLock()
	if len(s.activeSymbols) == 0 {
		s.mu.RUnlock()
		return // No active subscriptions, skip price updates
	}

	activeSymbolsList := make([]string, 0, len(s.activeSymbols))
	for symbol := range s.activeSymbols {
		activeSymbolsList = append(activeSymbolsList, symbol)
	}
	s.mu.RUnlock()

	// Only update a random subset of active symbols
	numToUpdate := mathRand.Intn(len(activeSymbolsList)) + 1
	if numToUpdate > len(activeSymbolsList) {
		numToUpdate = len(activeSymbolsList)
	}

	// Shuffle and pick random symbols from active ones
	mathRand.Shuffle(len(activeSymbolsList), func(i, j int) {
		activeSymbolsList[i], activeSymbolsList[j] = activeSymbolsList[j], activeSymbolsList[i]
	})

	// Get the assets to update
	allAssets := s.assetDataService.GetAllAssets()
	assetsToUpdate := make(map[string]*model.AssetQuote)

	for i := 0; i < numToUpdate; i++ {
		symbol := activeSymbolsList[i]
		if asset, exists := allAssets[symbol]; exists {
			newPrice := s.calculateNewPrice(asset)
			asset.UpdatePrice(newPrice)
			assetsToUpdate[symbol] = asset
		}
	}

	if len(assetsToUpdate) > 0 {
		s.notifySubscribers(assetsToUpdate)
	}
}

// Simulate the price oscillation
func (s *PriceOscillationService) calculateNewPrice(quote *model.AssetQuote) float64 {
	// Generate random oscillation between -1% and +1%
	oscillationPercent := (mathRand.Float64() - 0.5) * 2 * 0.01 // -0.01 to +0.01 (±1%)

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
		relevantAssets := make(map[string]*model.AssetQuote)
		for symbol, asset := range assets {
			if subscriber.symbols[symbol] {
				relevantAssets[symbol] = asset
			}
		}

		if len(relevantAssets) > 0 {
			select {
			case subscriber.channel <- relevantAssets:
			default:
				// Skip if subscriber channel is full to avoid blocking
				log.Printf("Subscriber %s channel full, skipping update", subscriber.id)
			}
		}
	}
}

func (s *PriceOscillationService) generateSubscriberID() string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to timestamp-based ID
		return hex.EncodeToString([]byte(time.Now().String()))[:16]
	}
	return hex.EncodeToString(bytes)
}

func (s *PriceOscillationService) mapToSlice(m map[string]bool) []string {
	slice := make([]string, 0, len(m))
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}

func (s *PriceOscillationService) getActiveSymbolsList() []string {
	slice := make([]string, 0, len(s.activeSymbols))
	for symbol := range s.activeSymbols {
		slice = append(slice, symbol)
	}
	return slice
}
