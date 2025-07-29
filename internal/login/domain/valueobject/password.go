package valueobject

import (
	"errors"
	"regexp"
	"unicode"
)

// Password represents a validated password value object
type Password struct {
	value string
}

// Password validation regex patterns
var (
	// hasUppercase checks for at least one uppercase letter
	hasUppercase = regexp.MustCompile(`[A-Z]`)
	// hasSpecialChar checks for at least one special character
	hasSpecialChar = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':\"\\|,.<>\/?~` + "`" + `]`)
)

// NewPassword creates a new Password value object with validation
func NewPassword(password string) (*Password, error) {
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	return &Password{
		value: password, // Store password as-is (hashing should be done at application layer)
	}, nil
}

// NewPasswordFromRepository creates a Password value object without validation
// This should only be used when reconstructing from trusted sources (database)
func NewPasswordFromRepository(password string) *Password {
	return &Password{
		value: password, // Assume password was already validated when stored
	}
}

// validatePassword performs comprehensive password validation
func validatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	// Check minimum length requirement first
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Check maximum length for security (prevent DOS attacks)
	if len(password) > 60 {
		return errors.New("password is too long (maximum 60 characters)")
	}

	// Check for at least one uppercase letter
	if !hasUppercase.MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLowercase := false
	for _, char := range password {
		if unicode.IsLower(char) {
			hasLowercase = true
			break
		}
	}
	if !hasLowercase {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := false
	for _, char := range password {
		if unicode.IsDigit(char) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	// Check for at least one special character
	if !hasSpecialChar.MatchString(password) {
		return errors.New("password must contain at least one special character (!@#$%^&*()_+-=[]{}:;\"'|,.<>?/~`)")
	}

	// Check for common weak patterns
	if err := checkWeakPatterns(password); err != nil {
		return err
	}

	return nil
}

// checkWeakPatterns checks for common weak password patterns
func checkWeakPatterns(password string) error {
	// Convert to lowercase for pattern checking
	lowerPassword := ""
	for _, char := range password {
		lowerPassword += string(unicode.ToLower(char))
	}

	// Common weak patterns
	weakPatterns := []string{
		"password", "123456", "qwerty", "abc123", "admin", "user",
		"login", "welcome", "changeme", "default", "guest",
		"12345678", "87654321", "qwertyui", "asdfghjk",
	}

	for _, pattern := range weakPatterns {
		if lowerPassword == pattern {
			return errors.New("password contains a common weak pattern")
		}
	}

	// Check for simple sequences
	if isSequential(password) {
		return errors.New("password cannot be a simple sequence")
	}

	return nil
}

// isSequential checks if password is a simple sequence like "12345678" or "abcdefgh"
func isSequential(password string) bool {
	if len(password) < 4 {
		return false
	}

	// Check for ascending sequences
	ascending := true
	for i := 1; i < len(password); i++ {
		if password[i] != password[i-1]+1 {
			ascending = false
			break
		}
	}

	// Check for descending sequences
	descending := true
	for i := 1; i < len(password); i++ {
		if password[i] != password[i-1]-1 {
			descending = false
			break
		}
	}

	return ascending || descending
}

// Value returns the password as a string
// Note: This should be used carefully and preferably only for hashing
func (p *Password) Value() string {
	return p.value
}

// Equals checks if two Password value objects are equal
func (p *Password) Equals(other *Password) bool {
	if other == nil {
		return false
	}
	return p.value == other.value
}

func (p *Password) EqualsString(other string) bool {
	return p.value == other
}

// Length returns the length of the password
func (p *Password) Length() int {
	return len(p.value)
}

// Strength returns a strength score from 1-5 based on password complexity
func (p *Password) Strength() int {
	score := 0

	// Length bonus
	if len(p.value) >= 8 {
		score++
	}
	if len(p.value) >= 12 {
		score++
	}

	// Character diversity bonus
	hasUpper := hasUppercase.MatchString(p.value)
	hasSpecial := hasSpecialChar.MatchString(p.value)

	var hasLower, hasDigit bool
	for _, char := range p.value {
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	characterTypes := 0
	if hasUpper {
		characterTypes++
	}
	if hasLower {
		characterTypes++
	}
	if hasDigit {
		characterTypes++
	}
	if hasSpecial {
		characterTypes++
	}

	// Add points based on character diversity
	if characterTypes >= 3 {
		score++
	}
	if characterTypes == 4 {
		score++
	}

	// Bonus for longer passwords
	if len(p.value) >= 16 {
		score++
	}

	// Ensure score is between 1 and 5
	if score < 1 {
		score = 1
	}
	if score > 5 {
		score = 5
	}

	return score
}

// IsValid checks if the current password is still valid (useful for cached instances)
func (p *Password) IsValid() bool {
	return validatePassword(p.value) == nil
}

// HasUppercase checks if password contains uppercase letters
func (p *Password) HasUppercase() bool {
	return hasUppercase.MatchString(p.value)
}

// HasSpecialChar checks if password contains special characters
func (p *Password) HasSpecialChar() bool {
	return hasSpecialChar.MatchString(p.value)
}

// HasDigit checks if password contains digits
func (p *Password) HasDigit() bool {
	for _, char := range p.value {
		if unicode.IsDigit(char) {
			return true
		}
	}
	return false
}

// HasLowercase checks if password contains lowercase letters
func (p *Password) HasLowercase() bool {
	for _, char := range p.value {
		if unicode.IsLower(char) {
			return true
		}
	}
	return false
}
