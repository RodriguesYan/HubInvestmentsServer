package repository

import "HubInvestments/internal/login/domain/model"

type ILoginRepository interface {
	GetUserByEmail(email string) (*model.User, error)
}
