package model

import (
	"HubInvestments/internal/login/domain/valueobject"
)

// User represents a user entity in the domain
type User struct {
	ID       string                `json:"id"`
	Email    *valueobject.Email    `json:"email"`
	Password *valueobject.Password `json:"-"` // Never serialize password
}

// NewUser creates a new User with validated email and password
func NewUser(id, email, password string) (*User, error) {
	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, err
	}

	passwordVO, err := valueobject.NewPassword(password)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Email:    emailVO,
		Password: passwordVO,
	}, nil
}

// NewUserFromRepository creates a User from database data without validation
// This should only be used when reconstructing users from trusted sources (database)
// where the data has already been validated upon creation
func NewUserFromRepository(id, email, password string) *User {
	// Create value objects without validation since data comes from database
	emailVO := valueobject.NewEmailFromRepository(email)
	passwordVO := valueobject.NewPasswordFromRepository(password)

	return &User{
		ID:       id,
		Email:    emailVO,
		Password: passwordVO,
	}
}

// GetEmailString returns the email as a string for compatibility
func (u *User) GetEmailString() string {
	if u.Email == nil {
		return ""
	}
	return u.Email.Value()
}

// GetPasswordString returns the password as a string for hashing/comparison
// This should only be used when necessary (e.g., for hashing)
func (u *User) GetPasswordString() string {
	if u.Password == nil {
		return ""
	}
	return u.Password.Value()
}

// ChangeEmail updates the user's email after validation
func (u *User) ChangeEmail(newEmail string) error {
	emailVO, err := valueobject.NewEmail(newEmail)
	if err != nil {
		return err
	}
	u.Email = emailVO
	return nil
}

// ChangePassword updates the user's password after validation
func (u *User) ChangePassword(newPassword string) error {
	passwordVO, err := valueobject.NewPassword(newPassword)
	if err != nil {
		return err
	}
	u.Password = passwordVO
	return nil
}
