package main

import (
	// formatting and printing values to the console.
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

type LoginModel struct {
	Email    string
	Password string
}

func Login(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Login success")
	decoder := json.NewDecoder(r.Body)
	var t LoginModel
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	log.Println(t.Email)
	log.Println(t.Password)

}

func main() {
	http.HandleFunc("/", Home)
	http.HandleFunc("/login", Login)

	err := http.ListenAndServe(portNum, nil)

	if err != nil {
		log.Fatal(err)
	}
}
