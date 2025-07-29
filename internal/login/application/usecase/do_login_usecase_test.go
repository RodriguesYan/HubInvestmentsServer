package usecase

import (
	"HubInvestments/internal/login/domain/model"
	"HubInvestments/internal/login/domain/valueobject"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type LoginRepositoryMock struct {
	mock.Mock
}

func (l *LoginRepositoryMock) GetUserByEmail(email string) (*model.User, error) {
	args := l.Called(email)
	return args.Get(0).(*model.User), args.Error(1)
}

func TestDoLoginUsecase_Execute_Success(t *testing.T) {
	// Arrange
	repo := &LoginRepositoryMock{}
	expectedData := &model.User{
		Email:    valueobject.NewEmailFromRepository("myemail@myemail.com"),
		ID:       "1",
		Password: valueobject.NewPasswordFromRepository("123456"),
	}
	repo.On("GetUserByEmail", "myemail@myemail.com").Return(expectedData, nil)
	usecase := NewDoLoginUsecase(repo)

	// Act
	result, err := usecase.Execute("myemail@myemail.com", "123456")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "1", result.ID)
	assert.Equal(t, "myemail@myemail.com", result.GetEmailString())
	repo.AssertExpectations(t)
}

func TestDoLoginUsecase_Execute_UserNotFound(t *testing.T) {
	// Arrange
	repo := &LoginRepositoryMock{}
	repo.On("GetUserByEmail", "notfound@myemail.com").Return((*model.User)(nil), errors.New("user not found"))
	usecase := NewDoLoginUsecase(repo)

	// Act
	result, err := usecase.Execute("notfound@myemail.com", "123456")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Equal(t, &model.User{}, result)
	repo.AssertExpectations(t)
}

func TestDoLoginUsecase_Execute_InvalidPassword(t *testing.T) {
	// Arrange
	repo := &LoginRepositoryMock{}
	userData := &model.User{
		Email:    valueobject.NewEmailFromRepository("myemail@myemail.com"),
		ID:       "1",
		Password: valueobject.NewPasswordFromRepository("correctpassword"),
	}
	repo.On("GetUserByEmail", "myemail@myemail.com").Return(userData, nil)
	usecase := NewDoLoginUsecase(repo)

	// Act
	result, err := usecase.Execute("myemail@myemail.com", "wrongpassword")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "invalid password", err.Error())
	assert.Equal(t, &model.User{}, result)
	repo.AssertExpectations(t)
}

func TestDoLoginUsecase_Execute_NilUser(t *testing.T) {
	// Arrange
	repo := &LoginRepositoryMock{}
	repo.On("GetUserByEmail", "myemail@myemail.com").Return((*model.User)(nil), nil)
	usecase := NewDoLoginUsecase(repo)

	// Act
	result, err := usecase.Execute("myemail@myemail.com", "123456")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Equal(t, &model.User{}, result)
	repo.AssertExpectations(t)
}
