package domain

import (
	"testing"
)

func TestPositionStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		ps       PositionStatus
		expected bool
	}{
		{"Valid ACTIVE", PositionStatusActive, true},
		{"Valid CLOSED", PositionStatusClosed, true},
		{"Valid PARTIAL", PositionStatusPartial, true},
		{"Invalid empty", PositionStatus(""), false},
		{"Invalid value", PositionStatus("INVALID"), false},
		{"Invalid lowercase", PositionStatus("active"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.IsValid(); got != tt.expected {
				t.Errorf("PositionStatus.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewPositionStatus(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  PositionStatus
		wantError bool
	}{
		{"Valid ACTIVE uppercase", "ACTIVE", PositionStatusActive, false},
		{"Valid ACTIVE lowercase", "active", PositionStatusActive, false},
		{"Valid ACTIVE with spaces", "  ACTIVE  ", PositionStatusActive, false},
		{"Valid CLOSED uppercase", "CLOSED", PositionStatusClosed, false},
		{"Valid CLOSED lowercase", "closed", PositionStatusClosed, false},
		{"Valid CLOSED with spaces", "  CLOSED  ", PositionStatusClosed, false},
		{"Valid PARTIAL uppercase", "PARTIAL", PositionStatusPartial, false},
		{"Valid PARTIAL lowercase", "partial", PositionStatusPartial, false},
		{"Valid PARTIAL with spaces", "  PARTIAL  ", PositionStatusPartial, false},
		{"Invalid empty", "", "", true},
		{"Invalid value", "INVALID", "", true},
		{"Valid mixed case", "Active", PositionStatusActive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPositionStatus(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewPositionStatus() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewPositionStatus() unexpected error: %v", err)
				return
			}

			if got != tt.expected {
				t.Errorf("NewPositionStatus() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPositionStatus_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		ps       PositionStatus
		expected bool
	}{
		{"ACTIVE status", PositionStatusActive, true},
		{"CLOSED status", PositionStatusClosed, false},
		{"PARTIAL status", PositionStatusPartial, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.IsActive(); got != tt.expected {
				t.Errorf("PositionStatus.IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPositionStatus_IsClosed(t *testing.T) {
	tests := []struct {
		name     string
		ps       PositionStatus
		expected bool
	}{
		{"ACTIVE status", PositionStatusActive, false},
		{"CLOSED status", PositionStatusClosed, true},
		{"PARTIAL status", PositionStatusPartial, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.IsClosed(); got != tt.expected {
				t.Errorf("PositionStatus.IsClosed() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPositionStatus_IsPartial(t *testing.T) {
	tests := []struct {
		name     string
		ps       PositionStatus
		expected bool
	}{
		{"ACTIVE status", PositionStatusActive, false},
		{"CLOSED status", PositionStatusClosed, false},
		{"PARTIAL status", PositionStatusPartial, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.IsPartial(); got != tt.expected {
				t.Errorf("PositionStatus.IsPartial() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPositionStatus_CanBeUpdated(t *testing.T) {
	tests := []struct {
		name     string
		ps       PositionStatus
		expected bool
	}{
		{"ACTIVE status", PositionStatusActive, true},
		{"CLOSED status", PositionStatusClosed, false},
		{"PARTIAL status", PositionStatusPartial, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.CanBeUpdated(); got != tt.expected {
				t.Errorf("PositionStatus.CanBeUpdated() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPositionStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		ps       PositionStatus
		expected string
	}{
		{"ACTIVE status", PositionStatusActive, "ACTIVE"},
		{"CLOSED status", PositionStatusClosed, "CLOSED"},
		{"PARTIAL status", PositionStatusPartial, "PARTIAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.String(); got != tt.expected {
				t.Errorf("PositionStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAllPositionStatuses(t *testing.T) {
	statuses := AllPositionStatuses()

	if len(statuses) != 3 {
		t.Errorf("AllPositionStatuses() expected 3 statuses, got %d", len(statuses))
	}

	expectedStatuses := map[PositionStatus]bool{
		PositionStatusActive:  true,
		PositionStatusClosed:  true,
		PositionStatusPartial: true,
	}

	for _, ps := range statuses {
		if !expectedStatuses[ps] {
			t.Errorf("AllPositionStatuses() contains unexpected status: %v", ps)
		}
	}
}
