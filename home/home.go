package home

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"HubInvestments/auth"

	"github.com/jmoiron/sqlx"
)

type AucAggregationModel struct {
	Id                  int     `json:"id" db:"id"`
	UserId              int     `json:"userId" db:"user_id"`
	AvailableBalance    float32 `json:"availableBalance" db:"available_balance"`
	FixedIncomeInvested float32 `json:"fixedIncomeInvested" db:"fixed_income_invested"`
	StocksInvested      float32 `json:"stocksInvested" db:"stocks_invested"`
	EtfsInvested        float32 `json:"etfsInvested" db:"etfs_invested"`
}

func GetAucAggregation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")

	userId, err := auth.VerifyToken(tokenString, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	fmt.Printf("caiu aqui 3")

	db, err := sqlx.Connect("postgres", "user=yanrodrigues dbname=yanrodrigues sslmode=disable password= host=localhost")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully Connected")
	}

	aggregation, err := db.Queryx("SELECT * FROM aucAggregations where user_id = $1", userId)

	if err != nil {
		log.Println("Caindo nesse erro aqui 1")
		log.Fatal(err)
	}

	var aucAggregations AucAggregationModel

	for aggregation.Next() {
		if err := aggregation.StructScan(&aucAggregations); err != nil {
			log.Println("Caindo nesse erro aqui 2")
			log.Fatal(err)
		}
	}

	result, err := json.Marshal(aucAggregations)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Println("My result", string(result))

	fmt.Fprint(w, string(result))
}
