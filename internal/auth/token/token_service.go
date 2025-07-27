package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

var secretKey = []byte("secret-key") //TODO: por essa key em um env file da vida

type ITokenService interface {
	CreateAndSignToken(userName string, userId string) (string, error)
	ValidateToken(tokenString string) (map[string]interface{}, error)
}

type TokenService struct{}

type TokenClaims map[string]interface{}

func NewTokenService() ITokenService {
	return &TokenService{}
}

func (s *TokenService) CreateAndSignToken(userName string, userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": userName,
			"userId":   userId,
			"exp":      time.Now().Add(time.Minute * 10).Unix(), //token expiration time = 1 min
		})

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *TokenService) ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := s.parseToken(tokenString)

	if err != nil {
		return nil, err
	}

	claims, err := validateToken(token)

	if err != nil {
		return nil, err
	}

	bla := TokenClaims(claims)

	return bla, nil
}

func (s *TokenService) parseToken(token string) (*jwt.Token, error) {
	token = token[len("Bearer "):]

	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	return jwtToken, err
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
