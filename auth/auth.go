package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var secretKey = []byte("secret-key") //TODO: por essa key em um env file da vida

type AuthService struct {
	jwtParser jwt.Parser
}

func NewAuthService(jwtParser jwt.Parser) *AuthService {
	return &AuthService{jwtParser: jwtParser}
}

//TODO: por tudo que é de jwt em um file apartado (auth_adapter ou algo do tipo)
//TODO: criar interface e tudo pra conseguir mockar em testes, ai esse cara aqui fica livre dessa lib

func (s *AuthService) VerifyToken(tokenString string, w http.ResponseWriter) (string, error) {
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")

		return "", errors.New("missing authorization header")
	}

	token, err := s.parseToken(tokenString)

	if err != nil {
		return "", err
	}

	claims, err := validateToken(token)

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

func (s *AuthService) parseToken(token string) (*jwt.Token, error) {
	token = token[len("Bearer "):]

	jwtToken, err := s.jwtParser.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	return jwtToken, err
}

func CreateToken(userName string, userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": userName,
			"userId":   userId,
			"exp":      time.Now().Add(time.Minute * 1).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
