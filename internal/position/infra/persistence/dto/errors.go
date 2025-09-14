package dto

import "errors"

var (
	ErrPositionNotFound      = errors.New("position not found")
	ErrDuplicatePosition     = errors.New("duplicate position for user and symbol")
	ErrInvalidPositionType   = errors.New("invalid position type")
	ErrInvalidPositionStatus = errors.New("invalid position status")
)
