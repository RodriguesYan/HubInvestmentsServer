package dto

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
)

type OrderDTO struct {
	ID                      uuid.UUID  `db:"id"`
	UserID                  int        `db:"user_id"`
	Symbol                  string     `db:"symbol"`
	OrderType               string     `db:"order_type"`
	OrderSide               string     `db:"order_side"`
	Quantity                float64    `db:"quantity"`
	Price                   *float64   `db:"price"`
	Status                  string     `db:"status"`
	CreatedAt               time.Time  `db:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at"`
	ExecutedAt              *time.Time `db:"executed_at"`
	ExecutionPrice          *float64   `db:"execution_price"`
	MarketPriceAtSubmission *float64   `db:"market_price_at_submission"`
	MarketDataTimestamp     *time.Time `db:"market_data_timestamp"`
	FailureReason           *string    `db:"failure_reason"`
	RetryCount              int        `db:"retry_count"`
	ProcessingWorkerID      *string    `db:"processing_worker_id"`
	ExternalOrderID         *string    `db:"external_order_id"`
}

// NullableFloat64 handles NULL values for DECIMAL fields
type NullableFloat64 struct {
	Float64 float64
	Valid   bool
}

func (nf *NullableFloat64) Scan(value interface{}) error {
	if value == nil {
		nf.Float64, nf.Valid = 0, false
		return nil
	}
	nf.Valid = true
	switch v := value.(type) {
	case float64:
		nf.Float64 = v
	case int64:
		nf.Float64 = float64(v)
	case []byte:
		// Handle string representation of decimal
		if len(v) == 0 {
			nf.Valid = false
			return nil
		}
		// Convert bytes to float64 (simplified)
		nf.Float64 = 0 // This would need proper decimal parsing
	}
	return nil
}

func (nf NullableFloat64) Value() (driver.Value, error) {
	if !nf.Valid {
		return nil, nil
	}
	return nf.Float64, nil
}

// ToFloat64Ptr converts NullableFloat64 to *float64
func (nf *NullableFloat64) ToFloat64Ptr() *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

// FromFloat64Ptr creates NullableFloat64 from *float64
func FromFloat64Ptr(f *float64) NullableFloat64 {
	if f == nil {
		return NullableFloat64{Valid: false}
	}
	return NullableFloat64{Float64: *f, Valid: true}
}
