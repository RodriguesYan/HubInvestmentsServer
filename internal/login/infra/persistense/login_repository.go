package persistense

import (
	"HubInvestments/internal/login/domain/model"
	"HubInvestments/internal/login/domain/repository"
	"HubInvestments/shared/infra/database"
	"fmt"
)

type LoginRepository struct {
	db database.Database
}

// userDTO represents the database structure for user data
type userDTO struct {
	ID       string `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
}

func NewLoginRepository(db database.Database) repository.ILoginRepository {
	return &LoginRepository{db: db}
}

func (l *LoginRepository) GetUserByEmail(email string) (*model.User, error) {
	query := "SELECT id, email, password FROM users WHERE email = $1"

	var userDB userDTO
	err := l.db.Get(&userDB, query, email)

	if err != nil {
		return nil, fmt.Errorf("user not found or database error: %w", err)
	}

	// Convert DTO to domain model without validation (data comes from trusted database)
	user := model.NewUserFromRepository(userDB.ID, userDB.Email, userDB.Password)

	return user, nil
}
