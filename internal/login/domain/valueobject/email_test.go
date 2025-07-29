package valueobject

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmail_ValidEmails(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.com",
		"user+tag@example.org",
		"user_name@example-domain.com",
		"test123@test123.com",
		"a@b.co",
		"long.email.address@very-long-domain-name.com",
		"user@subdomain.example.com",
		"test.email+tag@example.co.uk",
		"1234567890@example.com",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			emailVO, err := NewEmail(email)
			assert.NoError(t, err)
			assert.NotNil(t, emailVO)
			assert.Equal(t, strings.ToLower(email), emailVO.Value())
		})
	}
}

func TestNewEmail_InvalidEmails(t *testing.T) {
	invalidEmails := []struct {
		email    string
		errorMsg string
	}{
		{"", "email cannot be empty"},
		{"   ", "email cannot be empty"},
		{"plainaddress", "email must contain exactly one @ symbol"},
		{"@missingdomain.com", "email local part must be between 1 and 64 characters"},
		{"missing@.com", "invalid email format"},
		{"missing@domain", "invalid email format"},
		{"missing.domain@.com", "invalid email format"},
		{"user@", "email domain part must be between 1 and 253 characters"},
		{"user@@domain.com", "email must contain exactly one @ symbol"},
		{"user@domain@com", "email must contain exactly one @ symbol"},
		{"user..name@domain.com", "email cannot contain consecutive dots"},
		{".user@domain.com", "email local part cannot start or end with a dot"},
		{"user.@domain.com", "email local part cannot start or end with a dot"},
		{"user@domain..com", "email cannot contain consecutive dots"},
		{strings.Repeat("a", 65) + "@domain.com", "email local part must be between 1 and 64 characters"},
		{"user@" + strings.Repeat("a", 254) + ".com", "email address is too long (maximum 254 characters)"},
		{strings.Repeat("a", 255) + "@domain.com", "email address is too long (maximum 254 characters)"},
		{"user name@domain.com", "invalid email format"},
		{"user@domain .com", "invalid email format"},
	}

	for _, testCase := range invalidEmails {
		t.Run(testCase.email, func(t *testing.T) {
			emailVO, err := NewEmail(testCase.email)
			assert.Error(t, err)
			assert.Nil(t, emailVO)
			assert.Contains(t, err.Error(), testCase.errorMsg)
		})
	}
}

func TestEmail_Value(t *testing.T) {
	email := "Test.User@EXAMPLE.COM"
	emailVO, err := NewEmail(email)

	assert.NoError(t, err)
	assert.Equal(t, "test.user@example.com", emailVO.Value()) // Should be normalized to lowercase
}

func TestEmail_Equals(t *testing.T) {
	email1, _ := NewEmail("test@example.com")
	email2, _ := NewEmail("test@example.com")
	email3, _ := NewEmail("different@example.com")

	assert.True(t, email1.Equals(email2))
	assert.False(t, email1.Equals(email3))
	assert.False(t, email1.Equals(nil))
}

func TestEmail_Domain(t *testing.T) {
	testCases := []struct {
		email          string
		expectedDomain string
	}{
		{"test@example.com", "example.com"},
		{"user@subdomain.example.org", "subdomain.example.org"},
		{"simple@test.co.uk", "test.co.uk"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.email, func(t *testing.T) {
			emailVO, err := NewEmail(testCase.email)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedDomain, emailVO.Domain())
		})
	}
}

func TestEmail_LocalPart(t *testing.T) {
	testCases := []struct {
		email             string
		expectedLocalPart string
	}{
		{"test@example.com", "test"},
		{"user.name@example.com", "user.name"},
		{"user+tag@example.com", "user+tag"},
		{"123@example.com", "123"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.email, func(t *testing.T) {
			emailVO, err := NewEmail(testCase.email)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedLocalPart, emailVO.LocalPart())
		})
	}
}

func TestEmail_IsValid(t *testing.T) {
	validEmail, _ := NewEmail("test@example.com")
	assert.True(t, validEmail.IsValid())
}

func TestEmail_Normalization(t *testing.T) {
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
			emailVO, err := NewEmail(testCase.input)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, emailVO.Value())
		})
	}
}

func TestEmail_EdgeCases(t *testing.T) {
	t.Run("minimum valid email", func(t *testing.T) {
		emailVO, err := NewEmail("a@b.co")
		assert.NoError(t, err)
		assert.Equal(t, "a@b.co", emailVO.Value())
	})

	t.Run("maximum length local part", func(t *testing.T) {
		localPart := strings.Repeat("a", 64)
		email := localPart + "@example.com"
		emailVO, err := NewEmail(email)
		assert.NoError(t, err)
		assert.Equal(t, 64, len(emailVO.LocalPart()))
	})

	t.Run("email with numbers and hyphens", func(t *testing.T) {
		emailVO, err := NewEmail("test123@sub-domain.example.com")
		assert.NoError(t, err)
		assert.NotNil(t, emailVO)
	})
}
