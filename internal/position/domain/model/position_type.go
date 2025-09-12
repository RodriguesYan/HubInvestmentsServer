package domain

import (
	"errors"
	"strings"
)

type PositionType string

const (
	PositionTypeLong  PositionType = "LONG"  // Long position (bought shares)
	PositionTypeShort PositionType = "SHORT" // Short position (sold shares without owning)
)

func AllPositionTypes() []PositionType {
	return []PositionType{
		PositionTypeLong,
		PositionTypeShort,
	}
}

func (pt PositionType) IsValid() bool {
	for _, validType := range AllPositionTypes() {
		if pt == validType {
			return true
		}
	}
	return false
}

func (pt PositionType) String() string {
	return string(pt)
}

// NewPositionType creates a new PositionType from string
func NewPositionType(value string) (PositionType, error) {
	upperValue := strings.ToUpper(strings.TrimSpace(value))
	positionType := PositionType(upperValue)

	if !positionType.IsValid() {
		return "", errors.New("invalid position type: must be LONG or SHORT")
	}

	return positionType, nil
}

func (pt PositionType) IsLong() bool {
	return pt == PositionTypeLong
}

func (pt PositionType) IsShort() bool {
	return pt == PositionTypeShort
}
