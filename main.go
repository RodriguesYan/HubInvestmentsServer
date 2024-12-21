package main

import (
	"HubInvestments/home"
	"HubInvestments/login"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const portNum string = "localhost:8080"

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

	log.Println(t.Email)
	log.Println(t.Password)

	fmt.Fprint(w, token)
}

func main() {
	http.HandleFunc("/login", Login)
	http.HandleFunc("/getBalance", home.GetBalance)

	err := http.ListenAndServe(portNum, nil)

	if err != nil {
		log.Fatal(err)
	}
}
