package usecase

import (
	"context"
	"errors"

	domain "HubInvestments/internal/position/domain/model"

	"github.com/google/uuid"
)

// MockPositionRepositoryForNew is a mock implementation of IPositionRepository for new Position domain model
type MockPositionRepositoryForNew struct {
	positions        map[string]*domain.Position
	existsForUser    map[string]bool
	positionCounts   map[string]int
	totalInvestments map[string]float64
	shouldFailSave   bool
	shouldFailExists bool
	shouldFailFind   bool
	shouldFailUpdate bool
}

func NewMockPositionRepositoryForNew() *MockPositionRepositoryForNew {
	return &MockPositionRepositoryForNew{
		positions:        make(map[string]*domain.Position),
		existsForUser:    make(map[string]bool),
		positionCounts:   make(map[string]int),
		totalInvestments: make(map[string]float64),
	}
}

// Legacy method for compatibility
func (m *MockPositionRepositoryForNew) GetPositionsByUserId(userId string) ([]domain.AssetModel, error) {
	return nil, errors.New("legacy method not implemented in new mock")
}

func (m *MockPositionRepositoryForNew) FindByID(ctx context.Context, positionID uuid.UUID) (*domain.Position, error) {
	if m.shouldFailFind {
		return nil, errors.New("mock find error")
	}
	position, exists := m.positions[positionID.String()]
	if !exists {
		return nil, nil
	}
	return position, nil
}

func (m *MockPositionRepositoryForNew) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error) {
	if m.shouldFailFind {
		return nil, errors.New("mock find error")
	}
	var userPositions []*domain.Position
	for _, position := range m.positions {
		if position.UserID == userID {
			userPositions = append(userPositions, position)
		}
	}
	return userPositions, nil
}

func (m *MockPositionRepositoryForNew) FindByUserIDAndSymbol(ctx context.Context, userID uuid.UUID, symbol string) (*domain.Position, error) {
	if m.shouldFailFind {
		return nil, errors.New("mock find error")
	}
	for _, position := range m.positions {
		if position.UserID == userID && position.Symbol == symbol {
			return position, nil
		}
	}
	return nil, nil
}

func (m *MockPositionRepositoryForNew) FindActivePositions(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error) {
	if m.shouldFailFind {
		return nil, errors.New("mock find error")
	}
	var activePositions []*domain.Position
	for _, position := range m.positions {
		if position.UserID == userID && position.Status.CanBeUpdated() {
			activePositions = append(activePositions, position)
		}
	}
	return activePositions, nil
}

func (m *MockPositionRepositoryForNew) Save(ctx context.Context, position *domain.Position) error {
	if m.shouldFailSave {
		return errors.New("mock save error")
	}
	m.positions[position.ID.String()] = position
	return nil
}

func (m *MockPositionRepositoryForNew) Update(ctx context.Context, position *domain.Position) error {
	if m.shouldFailUpdate {
		return errors.New("mock update error")
	}
	m.positions[position.ID.String()] = position
	return nil
}

func (m *MockPositionRepositoryForNew) Delete(ctx context.Context, positionID uuid.UUID) error {
	delete(m.positions, positionID.String())
	return nil
}

func (m *MockPositionRepositoryForNew) ExistsForUser(ctx context.Context, userID uuid.UUID, symbol string) (bool, error) {
	if m.shouldFailExists {
		return false, errors.New("mock exists error")
	}
	key := userID.String() + ":" + symbol
	return m.existsForUser[key], nil
}

func (m *MockPositionRepositoryForNew) CountPositionsForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	return m.positionCounts[userID.String()], nil
}

func (m *MockPositionRepositoryForNew) GetTotalInvestmentForUser(ctx context.Context, userID uuid.UUID) (float64, error) {
	return m.totalInvestments[userID.String()], nil
}

// Helper methods for testing
func (m *MockPositionRepositoryForNew) SetExistsForUser(userID uuid.UUID, symbol string, exists bool) {
	key := userID.String() + ":" + symbol
	m.existsForUser[key] = exists
}

func (m *MockPositionRepositoryForNew) SetPositionCount(userID uuid.UUID, count int) {
	m.positionCounts[userID.String()] = count
}

func (m *MockPositionRepositoryForNew) SetTotalInvestment(userID uuid.UUID, amount float64) {
	m.totalInvestments[userID.String()] = amount
}

func (m *MockPositionRepositoryForNew) AddPosition(position *domain.Position) {
	m.positions[position.ID.String()] = position
}

func (m *MockPositionRepositoryForNew) GetPositionByID(positionID uuid.UUID) *domain.Position {
	return m.positions[positionID.String()]
}

func (m *MockPositionRepositoryForNew) GetPositionCount() int {
	return len(m.positions)
}

// MockPositionRepositoryLegacy is for legacy tests that only use the old interface
type MockPositionRepositoryLegacy struct {
	model []domain.AssetModel
	err   error
}

func (r MockPositionRepositoryLegacy) GetPositionsByUserId(userId string) ([]domain.AssetModel, error) {
	if r.err != nil {
		return []domain.AssetModel{}, r.err
	}
	return r.model, nil
}

// Legacy mock doesn't implement new methods - will cause compile error if used incorrectly
func (r MockPositionRepositoryLegacy) FindByID(ctx context.Context, positionID uuid.UUID) (*domain.Position, error) {
	return nil, errors.New("legacy mock doesn't support new methods")
}

func (r MockPositionRepositoryLegacy) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error) {
	return nil, errors.New("legacy mock doesn't support new methods")
}

func (r MockPositionRepositoryLegacy) FindByUserIDAndSymbol(ctx context.Context, userID uuid.UUID, symbol string) (*domain.Position, error) {
	return nil, errors.New("legacy mock doesn't support new methods")
}

func (r MockPositionRepositoryLegacy) FindActivePositions(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error) {
	return nil, errors.New("legacy mock doesn't support new methods")
}

func (r MockPositionRepositoryLegacy) Save(ctx context.Context, position *domain.Position) error {
	return errors.New("legacy mock doesn't support new methods")
}

func (r MockPositionRepositoryLegacy) Update(ctx context.Context, position *domain.Position) error {
	return errors.New("legacy mock doesn't support new methods")
}

func (r MockPositionRepositoryLegacy) Delete(ctx context.Context, positionID uuid.UUID) error {
	return errors.New("legacy mock doesn't support new methods")
}

func (r MockPositionRepositoryLegacy) ExistsForUser(ctx context.Context, userID uuid.UUID, symbol string) (bool, error) {
	return false, errors.New("legacy mock doesn't support new methods")
}

func (r MockPositionRepositoryLegacy) CountPositionsForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, errors.New("legacy mock doesn't support new methods")
}

func (r MockPositionRepositoryLegacy) GetTotalInvestmentForUser(ctx context.Context, userID uuid.UUID) (float64, error) {
	return 0, errors.New("legacy mock doesn't support new methods")
}
