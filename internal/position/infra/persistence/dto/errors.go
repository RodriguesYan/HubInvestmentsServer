package dto

import "errors"

// DTO validation errors
var (
	ErrInvalidPositionID   = errors.New("invalid position ID")
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidSymbol       = errors.New("invalid symbol")
	ErrNegativeQuantity    = errors.New("quantity cannot be negative")
	ErrNegativePrice       = errors.New("price cannot be negative")
	ErrNegativeInvestment  = errors.New("investment cannot be negative")
	ErrInvalidPositionType = errors.New("invalid position type")
	ErrInvalidStatus       = errors.New("invalid status")
)

// Repository errors
var (
	ErrPositionNotFound       = errors.New("position not found")
	ErrDuplicatePosition      = errors.New("duplicate position for user and symbol")
	ErrPositionNotUpdated     = errors.New("position was not updated")
	ErrDatabaseConnection     = errors.New("database connection error")
	ErrTransactionFailed      = errors.New("database transaction failed")
	ErrInvalidQuery           = errors.New("invalid database query")
	ErrDataIntegrityViolation = errors.New("data integrity violation")
)

// Mapping errors
var (
	ErrMappingFailed        = errors.New("failed to map domain model to DTO")
	ErrInvalidDomainModel   = errors.New("invalid domain model for mapping")
	ErrMissingRequiredField = errors.New("missing required field in mapping")
)
