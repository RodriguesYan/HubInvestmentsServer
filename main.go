package main

import (
	"HubInvestments/auth"
	"HubInvestments/auth/token"
	"HubInvestments/login"
	di "HubInvestments/pck"
	get_aggregation "HubInvestments/position"
	"log"
	"net/http"
)

// const portNum string = "localhost:8080"
const portNum string = "192.168.0.4:8080" //My home IP
// const portNum string = "192.168.0.48:8080" //Camila's home IP

func main() {
	tokenService := token.NewTokenService()
	aucService := auth.NewAuthService(tokenService)

	verifyToken := func(token string, w http.ResponseWriter) (string, error) {
		return aucService.VerifyToken(token, w)
	}

	container, err := di.NewContainer()

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/login", login.Login)
	http.HandleFunc("/getAucAggregationBalance", func(w http.ResponseWriter, r *http.Request) {
		get_aggregation.GetAucAggregation(w, r, verifyToken, container)
	})

	err = http.ListenAndServe(portNum, nil)
	if err != nil {
		log.Fatal(err)
	}
}
