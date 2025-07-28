package usecase

import (
	"HubInvestments/internal/login/domain/model"
	"HubInvestments/internal/login/domain/repository"
	"errors"
)

type IDoLoginUsecase interface {
	Execute(email string, password string) (*model.User, error)
}

type DoLoginUsecase struct {
	repo repository.ILoginRepository
}

func NewDoLoginUsecase(repo repository.ILoginRepository) IDoLoginUsecase {
	return &DoLoginUsecase{repo: repo}
}

func (u *DoLoginUsecase) Execute(email string, password string) (*model.User, error) {

	user, err := u.repo.GetUserByEmail(email)
	if err != nil {
		return &model.User{}, err
	}

	if user == nil {
		return &model.User{}, errors.New("user not found")
	}

	if user.Password == nil {
		return &model.User{}, errors.New("user password not found")
	}

	if !user.Password.EqualsString(password) {
		return &model.User{}, errors.New("invalid password")
	}

	return user, nil
}
