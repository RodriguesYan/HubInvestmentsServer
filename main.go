package main

import (
	"HubInvestments/auth"
	"HubInvestments/home"
	"HubInvestments/login"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// const portNum string = "localhost:8080"
// const portNum string = "192.168.0.172:8080"//My home IP
const portNum string = "192.168.0.48:8080" //Camila's home IP

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var t login.LoginModel
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	token, err := login.Login(t, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(token)

	// log.Println(t.Email)
	// log.Println(t.Password)

	data := map[string]string{
		"token": token,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Fprint(w, string(jsonData))
}

func main() {
	// Define the token verification function
	verifyToken := func(token string, w http.ResponseWriter) (string, error) {
		return auth.VerifyToken(token, w)
	}

	// Define the database connection function
	connectDB := func() (*sqlx.DB, error) {
		return sqlx.Connect("postgres", "user=yanrodrigues dbname=yanrodrigues sslmode=disable password= host=localhost")
	}

	http.HandleFunc("/login", Login)
	// Create a closure that captures the dependencies
	http.HandleFunc("/getAucAggregationBalance", func(w http.ResponseWriter, r *http.Request) {
		home.GetAucAggregation(w, r, verifyToken, connectDB)
	})

	err := http.ListenAndServe(portNum, nil)

	if err != nil {
		log.Fatal(err)
	}
}
