package auth

import (
	"HubInvestments/internal/auth/token"
	"errors"
	"fmt"
	"net/http"
)

type IAuthService interface {
	VerifyToken(tokenString string, w http.ResponseWriter) (string, error)
	CreateToken(userName string, userId string) (string, error)
}

type AuthService struct {
	tokenService token.ITokenService
}

func NewAuthService(tokenService token.ITokenService) IAuthService {
	return &AuthService{tokenService: tokenService}
}

func (s *AuthService) VerifyToken(tokenString string, w http.ResponseWriter) (string, error) {
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")

		return "", errors.New("missing authorization header")
	}

	claims, err := s.tokenService.ValidateToken(tokenString)

	if err != nil {
		return "", err
	}

	userId, _ := claims["userId"].(string)

	return userId, nil
}

func (s *AuthService) CreateToken(userName string, userId string) (string, error) {
	return s.tokenService.CreateAndSignToken(userName, userId)
}
