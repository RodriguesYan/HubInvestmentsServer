# Login Domain Value Objects

This package contains value objects for the login domain, implementing Domain-Driven Design (DDD) principles. Value objects encapsulate validation logic and ensure data integrity at the domain level.

## Email Value Object

The `Email` value object provides robust email validation and normalization.

### Features

- **RFC 5322 compliant regex validation**
- **Email normalization** (lowercase, whitespace trimming)
- **Length constraints** (RFC 5321 limits)
- **Domain and local part validation**
- **Comprehensive error messages**

### Validation Rules

- ✅ Must be a valid email format according to RFC 5322
- ✅ Maximum length: 254 characters (RFC 5321)
- ✅ Local part: 1-64 characters (before @)
- ✅ Domain part: 1-253 characters (after @)
- ✅ Cannot contain consecutive dots (..)
- ✅ Local part cannot start or end with a dot
- ✅ Must contain exactly one @ symbol

### Usage Examples

```go
// Create a valid email
email, err := valueobject.NewEmail("user@example.com")
if err != nil {
    // Handle validation error
    log.Printf("Invalid email: %v", err)
    return
}

// Access email value
emailString := email.Value() // "user@example.com"

// Get domain and local part
domain := email.Domain()     // "example.com"
localPart := email.LocalPart() // "user"

// Email normalization
email, _ := valueobject.NewEmail("  User@EXAMPLE.COM  ")
fmt.Println(email.Value()) // "user@example.com"

// Compare emails
email1, _ := valueobject.NewEmail("test@example.com")
email2, _ := valueobject.NewEmail("test@example.com")
areEqual := email1.Equals(email2) // true
```

### Common Validation Errors

```go
// Empty email
_, err := valueobject.NewEmail("")
// Error: "email cannot be empty"

// Invalid format
_, err := valueobject.NewEmail("invalid-email")
// Error: "invalid email format"

// Missing domain
_, err := valueobject.NewEmail("user@")
// Error: "invalid email format"

// Consecutive dots
_, err := valueobject.NewEmail("user..name@domain.com")
// Error: "email cannot contain consecutive dots"
```

## Password Value Object

The `Password` value object enforces strong password policies for security.

### Features

- **Comprehensive strength validation**
- **Security-focused design** (masked string representation)
- **Weak pattern detection**
- **Sequential pattern detection**
- **Password strength scoring** (1-5 scale)
- **Character type validation**

### Validation Rules

- ✅ Minimum 8 characters
- ✅ At least one uppercase letter (A-Z)
- ✅ At least one lowercase letter (a-z)
- ✅ At least one digit (0-9)
- ✅ At least one special character (!@#$%^&*()_+-=[]{}:;"'|,.<>?/~`)
- ✅ Maximum 128 characters (security limit)
- ❌ Common weak patterns (password, 123456, qwerty, etc.)
- ❌ Simple sequences (abcdefgh, 12345678, etc.)

### Usage Examples

```go
// Create a valid password
password, err := valueobject.NewPassword("MySecure123!")
if err != nil {
    // Handle validation error
    log.Printf("Invalid password: %v", err)
    return
}

// Access password value (use carefully, preferably for hashing only)
passwordString := password.Value() // "MySecure123!"

// Security: String representation is masked
fmt.Println(password.String()) // "***HIDDEN***"

// Check password characteristics
hasUpper := password.HasUppercase()   // true
hasLower := password.HasLowercase()   // true
hasDigit := password.HasDigit()       // true
hasSpecial := password.HasSpecialChar() // true

// Get password strength (1-5 scale)
strength := password.Strength() // 3 (for example)

// Get password length
length := password.Length() // 12

// Compare passwords
password1, _ := valueobject.NewPassword("Test123!")
password2, _ := valueobject.NewPassword("Test123!")
areEqual := password1.Equals(password2) // true
```

### Password Strength Scoring

The password strength is calculated on a scale of 1-5:

- **1**: Meets minimum requirements only
- **2**: Basic strength (8+ chars, all required types)
- **3**: Good strength (12+ chars, diverse characters)
- **4**: Strong (varied character types, good length)
- **5**: Very strong (16+ chars, maximum diversity)

### Common Validation Errors

```go
// Too short
_, err := valueobject.NewPassword("Test1!")
// Error: "password must be at least 8 characters long"

// Missing uppercase
_, err := valueobject.NewPassword("test123!")
// Error: "password must contain at least one uppercase letter"

// Missing special character
_, err := valueobject.NewPassword("Test1234")
// Error: "password must contain at least one special character"

// Weak pattern
_, err := valueobject.NewPassword("Password123!")
// Error: "password contains a common weak pattern"

// Sequential pattern
_, err := valueobject.NewPassword("Abcdefgh1!")
// Error: "password cannot be a simple sequence"
```

## Integration with User Model

The value objects are integrated with the User domain model:

```go
// Create user with validated email and password
user, err := model.NewUser("user123", "test@example.com", "SecurePass123!")
if err != nil {
    // Handle validation error from either email or password
    return
}

// Access validated data
email := user.GetEmailString()    // "test@example.com"
password := user.GetPasswordString() // "SecurePass123!" (use for hashing)

// Change email with validation
err = user.ChangeEmail("new@example.com")
if err != nil {
    // Handle email validation error
}

// Change password with validation
err = user.ChangePassword("NewSecurePass456@")
if err != nil {
    // Handle password validation error
}

// Access value object methods through the user
domain := user.Email.Domain()           // "example.com"
strength := user.Password.Strength()    // 1-5
```

## Testing

Both value objects come with comprehensive test suites:

```bash
# Run all value object tests
go test -v ./internal/login/domain/valueobject/

# Run specific tests
go test -v ./internal/login/domain/valueobject/ -run TestEmail
go test -v ./internal/login/domain/valueobject/ -run TestPassword

# Run user model tests
go test -v ./internal/login/domain/model/
```

## Security Considerations

### Email Security
- **Normalization**: Emails are automatically normalized to lowercase
- **Input validation**: Comprehensive format validation prevents injection
- **Length limits**: Prevents buffer overflow attacks

### Password Security
- **Strong validation**: Enforces secure password policies
- **Masked representation**: `String()` method returns masked value
- **Weak pattern detection**: Prevents common weak passwords
- **Length limits**: Prevents DoS attacks with overly long passwords

### Best Practices

1. **Never log passwords**: Always use the masked `String()` representation
2. **Hash passwords immediately**: Use `Value()` only for hashing purposes
3. **Validate early**: Validation happens at object creation, fail fast
4. **Use domain methods**: Leverage the value object methods for business logic

## Error Handling

All validation errors are descriptive and user-friendly:

```go
email, err := valueobject.NewEmail("invalid@")
if err != nil {
    // err.Error() contains user-friendly message
    // "invalid email format"
}

password, err := valueobject.NewPassword("weak")
if err != nil {
    // err.Error() contains specific requirement
    // "password must be at least 8 characters long"
}
```

## Value Object Principles

These value objects follow DDD principles:

- ✅ **Immutable**: Cannot be changed after creation
- ✅ **Self-validating**: Validation logic is encapsulated
- ✅ **Equality by value**: Two objects with same value are equal
- ✅ **No identity**: No unique identifier, only value matters
- ✅ **Descriptive errors**: Clear error messages for validation failures

This ensures data integrity and business rule enforcement at the domain level. 