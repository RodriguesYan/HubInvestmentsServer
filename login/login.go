package login

import (
	"HubInvestments/auth"
	"HubInvestments/auth/token"
	"encoding/json"
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

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var t LoginModel
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	db, err := sqlx.Connect("postgres", "user=yanrodrigues dbname=yanrodrigues sslmode=disable password= host=localhost")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error connecting to DB:", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
	} else {
		log.Println("Successfully Connected")
	}

	user, err := db.Queryx("SELECT id, email, password FROM users where email = $1", t.Email)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		fmt.Println("Error doing sql query in users table:", err)
	}

	var email string
	var password string
	var userId string

	for user.Next() {
		if err := user.Scan(&userId, &email, &password); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal(err)
			fmt.Println("Error reading sql response:", err)
		}
	}

	if err := user.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		fmt.Println("Error scanning data to variables:", err)
	}

	if len(email) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("user or password is wrong")
	}

	if t.Password != password {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("user or password is wrong")
	}

	authService := auth.NewAuthService(token.NewTokenService())

	tokenString, err := authService.CreateToken(t.Email, string(userId))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("user or password is wrong")
	}

	data := map[string]string{
		"token": tokenString,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, string(jsonData))
}
