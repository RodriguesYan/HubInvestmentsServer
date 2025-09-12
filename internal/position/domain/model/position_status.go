package domain

import (
	"errors"
	"strings"
)

type PositionStatus string

const (
	PositionStatusActive  PositionStatus = "ACTIVE"  // Position is currently held
	PositionStatusClosed  PositionStatus = "CLOSED"  // Position has been fully closed
	PositionStatusPartial PositionStatus = "PARTIAL" // Position partially closed (has remaining quantity)
)

func AllPositionStatuses() []PositionStatus {
	return []PositionStatus{
		PositionStatusActive,
		PositionStatusClosed,
		PositionStatusPartial,
	}
}

func (ps PositionStatus) IsValid() bool {
	for _, validStatus := range AllPositionStatuses() {
		if ps == validStatus {
			return true
		}
	}
	return false
}

func (ps PositionStatus) String() string {
	return string(ps)
}

func NewPositionStatus(value string) (PositionStatus, error) {
	upperValue := strings.ToUpper(strings.TrimSpace(value))
	positionStatus := PositionStatus(upperValue)

	if !positionStatus.IsValid() {
		return "", errors.New("invalid position status: must be ACTIVE, CLOSED, or PARTIAL")
	}

	return positionStatus, nil
}

func (ps PositionStatus) IsActive() bool {
	return ps == PositionStatusActive
}

func (ps PositionStatus) IsClosed() bool {
	return ps == PositionStatusClosed
}

func (ps PositionStatus) IsPartial() bool {
	return ps == PositionStatusPartial
}

func (ps PositionStatus) CanBeUpdated() bool {
	return ps == PositionStatusActive || ps == PositionStatusPartial
}
