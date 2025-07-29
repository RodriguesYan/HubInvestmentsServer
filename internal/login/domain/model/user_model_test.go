package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser_ValidData(t *testing.T) {
	testCases := []struct {
		name     string
		id       string
		email    string
		password string
	}{
		{
			name:     "valid user data",
			id:       "user123",
			email:    "test@example.com",
			password: "TestPass123!",
		},
		{
			name:     "user with complex email",
			id:       "user456",
			email:    "user.name+tag@subdomain.example.org",
			password: "ComplexPass456@",
		},
		{
			name:     "user with strong password",
			id:       "user789",
			email:    "simple@test.com",
			password: "VeryStrongP@ssw0rd2024!",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			user, err := NewUser(testCase.id, testCase.email, testCase.password)

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, testCase.id, user.ID)
			assert.Equal(t, testCase.email, user.GetEmailString())
			assert.Equal(t, testCase.password, user.GetPasswordString())
		})
	}
}

func TestNewUser_InvalidEmail(t *testing.T) {
	invalidEmails := []string{
		"",
		"invalid-email",
		"@missing-local.com",
		"missing-domain@",
		"user..name@domain.com",
		".user@domain.com",
		"user.@domain.com",
	}

	for _, email := range invalidEmails {
		t.Run("invalid email: "+email, func(t *testing.T) {
			user, err := NewUser("user123", email, "ValidPass123!")

			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), "email")
		})
	}
}

func TestNewUser_InvalidPassword(t *testing.T) {
	invalidPasswords := []struct {
		password string
		errorMsg string
	}{
		{"", "password cannot be empty"},
		{"short", "password must be at least 8 characters long"},
		{"password", "password must contain at least one uppercase letter"},
		{"PASSWORD", "password must contain at least one lowercase letter"},
		{"Password", "password must contain at least one digit"},
		{"Password1", "password must contain at least one special character"},
		{"password123", "password must contain at least one uppercase letter"},
		{"PASSWORD123", "password must contain at least one lowercase letter"},
		{"Password!", "password must contain at least one digit"},
	}

	for _, testCase := range invalidPasswords {
		t.Run("invalid password: "+testCase.password, func(t *testing.T) {
			user, err := NewUser("user123", "test@example.com", testCase.password)

			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), testCase.errorMsg)
		})
	}
}

func TestUser_GetEmailString(t *testing.T) {
	user, err := NewUser("user123", "Test@Example.COM", "TestPass123!")

	assert.NoError(t, err)
	// Email should be normalized to lowercase
	assert.Equal(t, "test@example.com", user.GetEmailString())
}

func TestUser_GetPasswordString(t *testing.T) {
	password := "TestPass123!"
	user, err := NewUser("user123", "test@example.com", password)

	assert.NoError(t, err)
	assert.Equal(t, password, user.GetPasswordString())
}

func TestUser_ChangeEmail_Valid(t *testing.T) {
	user, err := NewUser("user123", "old@example.com", "TestPass123!")
	assert.NoError(t, err)

	newEmail := "new@example.com"
	err = user.ChangeEmail(newEmail)

	assert.NoError(t, err)
	assert.Equal(t, newEmail, user.GetEmailString())
}

func TestUser_ChangeEmail_Invalid(t *testing.T) {
	user, err := NewUser("user123", "valid@example.com", "TestPass123!")
	assert.NoError(t, err)

	invalidEmail := "invalid-email"
	err = user.ChangeEmail(invalidEmail)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email must contain exactly one @ symbol")
	// Original email should remain unchanged
	assert.Equal(t, "valid@example.com", user.GetEmailString())
}

func TestUser_ChangePassword_Valid(t *testing.T) {
	user, err := NewUser("user123", "test@example.com", "OldPass123!")
	assert.NoError(t, err)

	newPassword := "NewPass456@"
	err = user.ChangePassword(newPassword)

	assert.NoError(t, err)
	assert.Equal(t, newPassword, user.GetPasswordString())
}

func TestUser_ChangePassword_Invalid(t *testing.T) {
	user, err := NewUser("user123", "test@example.com", "ValidPass123!")
	assert.NoError(t, err)

	invalidPassword := "weak"
	err = user.ChangePassword(invalidPassword)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password must be at least 8 characters long")
	// Original password should remain unchanged
	assert.Equal(t, "ValidPass123!", user.GetPasswordString())
}

func TestUser_NilValueObjects(t *testing.T) {
	// Test edge case where value objects might be nil
	user := &User{
		ID:       "test",
		Email:    nil,
		Password: nil,
	}

	assert.Equal(t, "", user.GetEmailString())
	assert.Equal(t, "", user.GetPasswordString())
}

func TestUser_ValueObjectIntegration(t *testing.T) {
	// Test that the user correctly integrates with value objects
	user, err := NewUser("user123", "test@example.com", "TestPass123!")
	assert.NoError(t, err)

	// Test that value objects maintain their properties
	assert.True(t, user.Email.IsValid())
	assert.True(t, user.Password.IsValid())
	assert.True(t, user.Password.HasUppercase())
	assert.True(t, user.Password.HasLowercase())
	assert.True(t, user.Password.HasDigit())
	assert.True(t, user.Password.HasSpecialChar())
	assert.Equal(t, "example.com", user.Email.Domain())
	assert.Equal(t, "test", user.Email.LocalPart())
	assert.GreaterOrEqual(t, user.Password.Strength(), 1)
	assert.LessOrEqual(t, user.Password.Strength(), 5)
}

func TestUser_EmailNormalization(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Test@Example.COM", "test@example.com"},
		{"  user@domain.com  ", "user@domain.com"},
		{"USER+TAG@DOMAIN.ORG", "user+tag@domain.org"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			user, err := NewUser("user123", testCase.input, "TestPass123!")
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, user.GetEmailString())
		})
	}
}
