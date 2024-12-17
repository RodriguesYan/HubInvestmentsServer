package login

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type LoginModel struct {
	Email    string
	Password string
}

func Login(loginModel LoginModel) (string, error) {
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

	user, err := db.Queryx("SELECT email FROM users")

	if err != nil {
		log.Fatal(err)
		return "", err
	}

	var email string
	for user.Next() {
		if err := user.Scan(&email); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", email)
	}
	if err := user.Err(); err != nil {
		log.Fatal(err)
	}

	log.Println("User: ", email)

	return "Login success", nil
	//validar se password esta correto

	//retornar token
}
