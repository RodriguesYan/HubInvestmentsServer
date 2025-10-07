package command

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

// parseUserIDToUUID converts a user ID string to UUID.
// Supports both UUID format strings and integer strings (for backward compatibility).
// Integer user IDs (e.g., "1") are converted to UUID format: 00000000-0000-0000-0000-000000000001
func parseUserIDToUUID(userIDStr string) (uuid.UUID, error) {
	// First, try parsing as a direct UUID
	parsedUUID, err := uuid.Parse(userIDStr)
	if err == nil {
		return parsedUUID, nil
	}

	// If UUID parsing fails, try treating it as an integer and convert to UUID format
	userIDInt, scanErr := strconv.Atoi(userIDStr)
	if scanErr == nil {
		// Convert integer to UUID format: 00000000-0000-0000-0000-000000000001
		uuidStr := fmt.Sprintf("00000000-0000-0000-0000-%012d", userIDInt)
		return uuid.Parse(uuidStr)
	}

	return uuid.UUID{}, fmt.Errorf("invalid user ID format: %s", userIDStr)
}
