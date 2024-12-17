package main

import (
	// formatting and printing values to the console.
	"HubInvestments/login"
	"encoding/json"
	"fmt"
	"log"      // logging messages to the console.
	"net/http" // Used for build HTTP servers and clients.
)

const portNum string = ":8080"

func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage")
	fmt.Print(r.Body)
}

func Login(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t login.LoginModel
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	message, err := login.Login(t)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(message)

	log.Println(t.Email)
	log.Println(t.Password)
}

func Login2(w http.ResponseWriter, r *http.Request) {
	x := login.LoginModel{}
	err := json.NewDecoder(r.Body).Decode(&x)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Email: %s\nPassword: %s", x.Email, x.Password)
}

func main() {
	http.HandleFunc("/login", Login)

	err := http.ListenAndServe(portNum, nil)

	if err != nil {
		log.Fatal(err)
	}
}
