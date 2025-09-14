package persistence

import (
	domain "HubInvestments/internal/position/domain/model"
	repository "HubInvestments/internal/position/domain/repository"
	"HubInvestments/internal/position/infra/persistence/dto"
	"HubInvestments/shared/infra/database"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type PositionRepository struct {
	db     database.Database
	mapper *dto.PositionMapper
}

// NewPositionRepository creates a new position repository using the database abstraction
func NewPositionRepository(db database.Database) repository.PositionRepository {
	return &PositionRepository{
		db:     db,
		mapper: dto.NewPositionMapper(),
	}
}

// Position domain model methods - Full implementation with yanrodrigues.positions_v2 table

func (r *PositionRepository) FindByID(ctx context.Context, positionID uuid.UUID) (*domain.Position, error) {
	query := `
		SELECT id, user_id, symbol, quantity, average_price, total_investment, 
		       current_price, market_value, unrealized_pnl, unrealized_pnl_pct,
		       position_type, status, created_at, updated_at, last_trade_at
		FROM yanrodrigues.positions_v2 
		WHERE id = $1`

	var positionDTO dto.PositionDTO
	err := r.db.Get(&positionDTO, query, positionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("position %s not found: %w", positionID, dto.ErrPositionNotFound)
		}
		return nil, fmt.Errorf("failed to find position by ID: %w", err)
	}

	return r.mapper.ToDomain(&positionDTO)
}

func (r *PositionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error) {
	query := `
		SELECT id, user_id, symbol, quantity, average_price, total_investment,
		       current_price, market_value, unrealized_pnl, unrealized_pnl_pct,
		       position_type, status, created_at, updated_at, last_trade_at
		FROM yanrodrigues.positions_v2 
		WHERE user_id = $1
		ORDER BY created_at DESC`

	var positionDTOs []*dto.PositionDTO
	err := r.db.Select(&positionDTOs, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find positions for user %s: %w", userID, err)
	}

	return r.mapper.ToDomainList(positionDTOs)
}

func (r *PositionRepository) FindByUserIDAndSymbol(ctx context.Context, userID uuid.UUID, symbol string) (*domain.Position, error) {
	query := `
		SELECT id, user_id, symbol, quantity, average_price, total_investment,
		       current_price, market_value, unrealized_pnl, unrealized_pnl_pct,
		       position_type, status, created_at, updated_at, last_trade_at
		FROM yanrodrigues.positions_v2 
		WHERE user_id = $1 AND symbol = $2`

	var positionDTO dto.PositionDTO
	err := r.db.Get(&positionDTO, query, userID, symbol)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("position for user %s and symbol %s not found: %w", userID, symbol, dto.ErrPositionNotFound)
		}
		return nil, fmt.Errorf("failed to find position by user ID and symbol: %w", err)
	}

	return r.mapper.ToDomain(&positionDTO)
}

func (r *PositionRepository) FindActivePositions(ctx context.Context, userID uuid.UUID) ([]*domain.Position, error) {
	query := `
		SELECT id, user_id, symbol, quantity, average_price, total_investment,
		       current_price, market_value, unrealized_pnl, unrealized_pnl_pct,
		       position_type, status, created_at, updated_at, last_trade_at
		FROM yanrodrigues.positions_v2 
		WHERE user_id = $1 AND status IN ('ACTIVE', 'PARTIAL')
		ORDER BY created_at DESC`

	var positionDTOs []*dto.PositionDTO
	err := r.db.Select(&positionDTOs, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find active positions for user %s: %w", userID, err)
	}

	return r.mapper.ToDomainList(positionDTOs)
}

func (r *PositionRepository) Save(ctx context.Context, position *domain.Position) error {
	positionDTO, err := r.mapper.CreateDTOForInsert(position)
	if err != nil {
		return fmt.Errorf("failed to convert position to DTO: %w", err)
	}

	query := `
		INSERT INTO yanrodrigues.positions_v2 (
			id, user_id, symbol, quantity, average_price, total_investment,
			current_price, market_value, unrealized_pnl, unrealized_pnl_pct,
			position_type, status, created_at, updated_at, last_trade_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)`

	_, err = r.db.Exec(query,
		positionDTO.ID, positionDTO.UserID, positionDTO.Symbol,
		positionDTO.Quantity, positionDTO.AveragePrice, positionDTO.TotalInvestment,
		positionDTO.CurrentPrice, positionDTO.MarketValue, positionDTO.UnrealizedPnL,
		positionDTO.UnrealizedPnLPct, positionDTO.PositionType, positionDTO.Status,
		positionDTO.CreatedAt, positionDTO.UpdatedAt, positionDTO.LastTradeAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique_user_symbol") {
			return fmt.Errorf("position already exists for user %s and symbol %s: %w",
				positionDTO.UserID, positionDTO.Symbol, dto.ErrDuplicatePosition)
		}
		return fmt.Errorf("failed to save position: %w", err)
	}

	return nil
}

func (r *PositionRepository) Update(ctx context.Context, position *domain.Position) error {
	positionDTO, err := r.mapper.CreateDTOForUpdate(position)
	if err != nil {
		return fmt.Errorf("failed to convert position to DTO: %w", err)
	}

	query := `
		UPDATE yanrodrigues.positions_v2 SET
			quantity = $1,
			average_price = $2,
			total_investment = $3,
			current_price = $4,
			market_value = $5,
			unrealized_pnl = $6,
			unrealized_pnl_pct = $7,
			status = $8,
			updated_at = $9,
			last_trade_at = $10
		WHERE id = $11`

	result, err := r.db.Exec(query,
		positionDTO.Quantity, positionDTO.AveragePrice, positionDTO.TotalInvestment,
		positionDTO.CurrentPrice, positionDTO.MarketValue, positionDTO.UnrealizedPnL,
		positionDTO.UnrealizedPnLPct, positionDTO.Status, positionDTO.UpdatedAt,
		positionDTO.LastTradeAt, positionDTO.ID)
	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("position %s not found for update: %w", positionDTO.ID, dto.ErrPositionNotFound)
	}

	return nil
}

func (r *PositionRepository) Delete(ctx context.Context, positionID uuid.UUID) error {
	query := `DELETE FROM yanrodrigues.positions_v2 WHERE id = $1`

	result, err := r.db.Exec(query, positionID)
	if err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("position %s not found for deletion: %w", positionID, dto.ErrPositionNotFound)
	}

	return nil
}

func (r *PositionRepository) ExistsForUser(ctx context.Context, userID uuid.UUID, symbol string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0 
		FROM yanrodrigues.positions_v2 
		WHERE user_id = $1 AND symbol = $2`

	var exists bool
	err := r.db.Get(&exists, query, userID, symbol)
	if err != nil {
		return false, fmt.Errorf("failed to check position existence: %w", err)
	}

	return exists, nil
}

func (r *PositionRepository) CountPositionsForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM yanrodrigues.positions_v2 
		WHERE user_id = $1`

	var count int
	err := r.db.Get(&count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count positions for user: %w", err)
	}

	return count, nil
}

func (r *PositionRepository) GetTotalInvestmentForUser(ctx context.Context, userID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(total_investment), 0) 
		FROM yanrodrigues.positions_v2 
		WHERE user_id = $1 AND status IN ('ACTIVE', 'PARTIAL')`

	var totalInvestment float64
	err := r.db.Get(&totalInvestment, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get total investment for user: %w", err)
	}

	return totalInvestment, nil
}
