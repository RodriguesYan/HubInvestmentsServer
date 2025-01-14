package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
)

type IAuthService interface {
	VerifyToken(tokenString string, w http.ResponseWriter) (string, error)
	CreateToken(userName string, userId string) (string, error)
}

type AuthService struct {
	tokenService ITokenService
}

func NewAuthService(tokenService ITokenService) IAuthService {
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

func validateToken(token *jwt.Token) (jwt.MapClaims, error) {
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

func (s *AuthService) CreateToken(userName string, userId string) (string, error) {
	return s.tokenService.CreateAndSignToken(userName, userId)
}
