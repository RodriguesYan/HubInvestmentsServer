package login

import (
	"HubInvestments/auth"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type LoginModel struct {
	Email    string
	Password string
}

func Login(loginModel LoginModel, w http.ResponseWriter) (string, error) {
	db, err := sqlx.Connect("postgres", "user=yanrodrigues dbname=yanrodrigues sslmode=disable password= host=localhost")
	if err != nil {
		return "", err
	}
	//validar se email existe
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
		return "", err
	} else {
		log.Println("Successfully Connected")
	}

	user, err := db.Queryx("SELECT email, password FROM users where email = $1", loginModel.Email)

	if err != nil {
		log.Fatal(err)
		return "", err
	}

	var email string
	var password string

	for user.Next() {
		if err := user.Scan(&email, &password); err != nil {
			log.Fatal(err)
			return "", err
		}
		fmt.Printf("%s", email)
		fmt.Printf("%s", password)
	}

	if err := user.Err(); err != nil {
		log.Fatal(err)
		return "", err
	}

	if len(email) == 0 {
		return "", errors.New("user or password is wrong")
	}

	if loginModel.Password != password {
		return "", errors.New("user or password is wrong")
	}

	tokenString, err := auth.CreateToken(loginModel.Email)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return "", errors.New("user or password is wrong")
	}

	w.WriteHeader(http.StatusOK)

	return tokenString, nil
}
