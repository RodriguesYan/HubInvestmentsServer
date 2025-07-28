package valueobject

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPassword_ValidPasswords(t *testing.T) {
	validPasswords := []string{
		"Password1!",                        // Basic requirements met
		"MySecure123@",                      // All character types
		"Complex$Pass9",                     // Mixed case with special char and digit
		"Test123!@#",                        // Multiple special characters
		"SuperStr0ng&Secure",                // Longer password
		"8CharsA!",                          // Minimum length with all requirements
		"VeryLongPasswordWith123!@#$%^&*()", // Very long password
		"Mix3d$Case",                        // Another valid combination
		"Anoth3r!Valid",                     // Valid pattern
		"Secure2024#Pass",                   // Year with special char
	}

	for _, password := range validPasswords {
		t.Run(password, func(t *testing.T) {
			passwordVO, err := NewPassword(password)
			assert.NoError(t, err)
			assert.NotNil(t, passwordVO)
			assert.Equal(t, password, passwordVO.Value())
		})
	}
}

func TestNewPassword_InvalidPasswords(t *testing.T) {
	invalidPasswords := []struct {
		password string
		errorMsg string
	}{
		{"", "password cannot be empty"},
		{"short", "password must be at least 8 characters long"},
		{"1234567", "password must be at least 8 characters long"},
		{"password", "password must contain at least one uppercase letter"},
		{"PASSWORD", "password must contain at least one lowercase letter"},
		{"Password", "password must contain at least one digit"},
		{"Password1", "password must contain at least one special character"},
		{"password1!", "password must contain at least one uppercase letter"},
		{"PASSWORD1!", "password must contain at least one lowercase letter"},
		{"Password!", "password must contain at least one digit"},
		{"Password1", "password must contain at least one special character"},
		{"password", "password contains a common weak pattern"},
		{"123456", "password must be at least 8 characters long"},
		{"qwerty", "password must be at least 8 characters long"},
		{"12345678", "password contains a common weak pattern"},
		{"Password1234", "password must contain at least one special character"},
		{"abcdefgh", "password cannot be a simple sequence"},
		{"87654321", "password cannot be a simple sequence"},
		{strings.Repeat("a", 129), "password is too long (maximum 128 characters)"},
	}

	for _, testCase := range invalidPasswords {
		t.Run(testCase.password, func(t *testing.T) {
			passwordVO, err := NewPassword(testCase.password)
			assert.Error(t, err)
			assert.Nil(t, passwordVO)
			assert.Contains(t, err.Error(), testCase.errorMsg)
		})
	}
}

func TestPassword_Value(t *testing.T) {
	password := "TestPass123!"
	passwordVO, err := NewPassword(password)

	assert.NoError(t, err)
	assert.Equal(t, password, passwordVO.Value())
}

func TestPassword_Equals(t *testing.T) {
	password1, _ := NewPassword("TestPass123!")
	password2, _ := NewPassword("TestPass123!")
	password3, _ := NewPassword("DifferentPass123!")

	assert.True(t, password1.Equals(password2))
	assert.False(t, password1.Equals(password3))
	assert.False(t, password1.Equals(nil))
}

func TestPassword_Length(t *testing.T) {
	testCases := []struct {
		password       string
		expectedLength int
	}{
		{"TestPass123!", 12},
		{"Minimum8!", 9},
		{"VeryLongPasswordWith123!@#", 26},
	}

	for _, testCase := range testCases {
		t.Run(testCase.password, func(t *testing.T) {
			passwordVO, err := NewPassword(testCase.password)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedLength, passwordVO.Length())
		})
	}
}

func TestPassword_Strength(t *testing.T) {
	testCases := []struct {
		password        string
		expectedMinimum int
		expectedMaximum int
	}{
		{"TestPass123!", 3, 5},                      // Good password
		{"Minimum8!", 2, 4},                         // Shorter but meets requirements
		{"VeryLongPasswordWith123!@#$%^&*()", 5, 5}, // Very strong
		{"Valid123!", 1, 3},                         // Meets minimum requirements (8 chars)
	}

	for _, testCase := range testCases {
		t.Run(testCase.password, func(t *testing.T) {
			passwordVO, err := NewPassword(testCase.password)
			assert.NoError(t, err)
			strength := passwordVO.Strength()
			assert.GreaterOrEqual(t, strength, testCase.expectedMinimum)
			assert.LessOrEqual(t, strength, testCase.expectedMaximum)
			assert.GreaterOrEqual(t, strength, 1)
			assert.LessOrEqual(t, strength, 5)
		})
	}
}

func TestPassword_HasUppercase(t *testing.T) {
	passwordWithUpper, _ := NewPassword("TestPass123!")
	assert.True(t, passwordWithUpper.HasUppercase())
}

func TestPassword_HasSpecialChar(t *testing.T) {
	passwordWithSpecial, _ := NewPassword("TestPass123!")
	assert.True(t, passwordWithSpecial.HasSpecialChar())
}

func TestPassword_HasDigit(t *testing.T) {
	passwordWithDigit, _ := NewPassword("TestPass123!")
	assert.True(t, passwordWithDigit.HasDigit())
}

func TestPassword_HasLowercase(t *testing.T) {
	passwordWithLower, _ := NewPassword("TestPass123!")
	assert.True(t, passwordWithLower.HasLowercase())
}

func TestPassword_IsValid(t *testing.T) {
	validPassword, _ := NewPassword("TestPass123!")
	assert.True(t, validPassword.IsValid())
}

func TestPassword_WeakPatterns(t *testing.T) {
	weakPatterns := []string{
		"Password123!", // Contains "password"
		"Admin123!",    // Contains "admin"
		"User123!",     // Contains "user"
		"Login123!",    // Contains "login"
		"Welcome123!",  // Contains "welcome"
		"Qwerty123!",   // Contains "qwerty"
	}

	for _, password := range weakPatterns {
		t.Run(password, func(t *testing.T) {
			passwordVO, err := NewPassword(password)
			assert.Error(t, err)
			assert.Nil(t, passwordVO)
			assert.Contains(t, err.Error(), "password contains a common weak pattern")
		})
	}
}

func TestPassword_SequentialPatterns(t *testing.T) {
	sequentialPatterns := []string{
		"Abcdefgh1!", // Sequential letters
		"12345678!A", // Sequential numbers
		"87654321!A", // Reverse sequential numbers
	}

	for _, password := range sequentialPatterns {
		t.Run(password, func(t *testing.T) {
			passwordVO, err := NewPassword(password)
			assert.Error(t, err)
			assert.Nil(t, passwordVO)
			assert.Contains(t, err.Error(), "password cannot be a simple sequence")
		})
	}
}

func TestPassword_AllCharacterTypes(t *testing.T) {
	// Test that password validation requires all character types
	testCases := []struct {
		password string
		missing  string
	}{
		{"lowercase123!", "uppercase"},
		{"UPPERCASE123!", "lowercase"},
		{"NoNumbers!", "digit"},
		{"NoSpecialChars123", "special character"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.password+" missing "+testCase.missing, func(t *testing.T) {
			passwordVO, err := NewPassword(testCase.password)
			assert.Error(t, err)
			assert.Nil(t, passwordVO)
		})
	}
}

func TestPassword_SpecialCharacters(t *testing.T) {
	// Test various special characters
	specialChars := []string{
		"TestPass1!",
		"TestPass1@",
		"TestPass1#",
		"TestPass1$",
		"TestPass1%",
		"TestPass1^",
		"TestPass1&",
		"TestPass1*",
		"TestPass1(",
		"TestPass1)",
		"TestPass1_",
		"TestPass1+",
		"TestPass1-",
		"TestPass1=",
		"TestPass1[",
		"TestPass1]",
		"TestPass1{",
		"TestPass1}",
		"TestPass1;",
		"TestPass1:",
		"TestPass1'",
		"TestPass1\"",
		"TestPass1|",
		"TestPass1,",
		"TestPass1.",
		"TestPass1<",
		"TestPass1>",
		"TestPass1/",
		"TestPass1?",
		"TestPass1~",
		"TestPass1`",
	}

	for _, password := range specialChars {
		t.Run(password, func(t *testing.T) {
			passwordVO, err := NewPassword(password)
			assert.NoError(t, err)
			assert.NotNil(t, passwordVO)
			assert.True(t, passwordVO.HasSpecialChar())
		})
	}
}

func TestPassword_EdgeCases(t *testing.T) {
	t.Run("exactly 8 characters", func(t *testing.T) {
		password := "Test123!"
		passwordVO, err := NewPassword(password)
		assert.NoError(t, err)
		assert.Equal(t, 8, passwordVO.Length())
	})

	t.Run("maximum length", func(t *testing.T) {
		password := strings.Repeat("A", 120) + "test123!"
		passwordVO, err := NewPassword(password)
		assert.NoError(t, err)
		assert.Equal(t, 128, passwordVO.Length())
	})

	t.Run("Unicode characters", func(t *testing.T) {
		password := "TestPass123!αβγ" // Greek letters
		passwordVO, err := NewPassword(password)
		assert.NoError(t, err)
		assert.NotNil(t, passwordVO)
	})
}
