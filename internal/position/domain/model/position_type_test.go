package domain

import (
	"testing"
)

func TestPositionType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		pt       PositionType
		expected bool
	}{
		{"Valid LONG", PositionTypeLong, true},
		{"Valid SHORT", PositionTypeShort, true},
		{"Invalid empty", PositionType(""), false},
		{"Invalid value", PositionType("INVALID"), false},
		{"Invalid lowercase", PositionType("long"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pt.IsValid(); got != tt.expected {
				t.Errorf("PositionType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewPositionType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  PositionType
		wantError bool
	}{
		{"Valid LONG uppercase", "LONG", PositionTypeLong, false},
		{"Valid LONG lowercase", "long", PositionTypeLong, false},
		{"Valid LONG with spaces", "  LONG  ", PositionTypeLong, false},
		{"Valid SHORT uppercase", "SHORT", PositionTypeShort, false},
		{"Valid SHORT lowercase", "short", PositionTypeShort, false},
		{"Valid SHORT with spaces", "  SHORT  ", PositionTypeShort, false},
		{"Invalid empty", "", "", true},
		{"Invalid value", "INVALID", "", true},
		{"Valid mixed case", "Long", PositionTypeLong, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPositionType(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewPositionType() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewPositionType() unexpected error: %v", err)
				return
			}

			if got != tt.expected {
				t.Errorf("NewPositionType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPositionType_IsLong(t *testing.T) {
	tests := []struct {
		name     string
		pt       PositionType
		expected bool
	}{
		{"LONG position", PositionTypeLong, true},
		{"SHORT position", PositionTypeShort, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pt.IsLong(); got != tt.expected {
				t.Errorf("PositionType.IsLong() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPositionType_IsShort(t *testing.T) {
	tests := []struct {
		name     string
		pt       PositionType
		expected bool
	}{
		{"LONG position", PositionTypeLong, false},
		{"SHORT position", PositionTypeShort, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pt.IsShort(); got != tt.expected {
				t.Errorf("PositionType.IsShort() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPositionType_String(t *testing.T) {
	tests := []struct {
		name     string
		pt       PositionType
		expected string
	}{
		{"LONG position", PositionTypeLong, "LONG"},
		{"SHORT position", PositionTypeShort, "SHORT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pt.String(); got != tt.expected {
				t.Errorf("PositionType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAllPositionTypes(t *testing.T) {
	types := AllPositionTypes()

	if len(types) != 2 {
		t.Errorf("AllPositionTypes() expected 2 types, got %d", len(types))
	}

	expectedTypes := map[PositionType]bool{
		PositionTypeLong:  true,
		PositionTypeShort: true,
	}

	for _, pt := range types {
		if !expectedTypes[pt] {
			t.Errorf("AllPositionTypes() contains unexpected type: %v", pt)
		}
	}
}
