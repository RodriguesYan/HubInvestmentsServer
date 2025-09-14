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
