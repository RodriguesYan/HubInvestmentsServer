package home

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"

	"HubInvestments/auth"

	"github.com/jmoiron/sqlx"
)

type AssetsModel struct {
	Symbol       string  `json:"symbol" db:"symbol"`
	Quantity     float32 `json:"quantity" db:"quantity"`
	AveragePrice float32 `json:"averagePrice" db:"average_price"`
	LastPrice    float32 `json:"currentPrice" db:"current_price"`
}

type PositionAggregationModel struct {
	Category      int           `json:"category" db:"category"`
	TotalInvested float32       `json:"totalInvested" db:"total_invested"`
	CurrentTotal  float32       `json:"currentTotal" db:"current_total"`
	Pnl           float32       `json:"pnl" db:"pnl"`
	PnlPercentage float32       `json:"pnlPercentage" db:"pnl_percentage"`
	Assets        []AssetsModel `json:"assets"`
}

type AucAggregationModel struct {
	TotalBalance        float32                    `json:"totalBalance" db:"total_balance"`
	PositionAggregation []PositionAggregationModel `json:"positionAggregation"`
}

func GetAucAggregation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")

	userId, err := auth.VerifyToken(tokenString, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

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

	aggregation, err := db.Queryx(`
		SELECT 	i.symbol, 
				p.average_price, 
				p.quantity, 
				i.category, 
				i.last_price,
				b.current_value
		FROM positions p 
		join instruments i on p.instrument_id = i.id 
		join balance b on p.user_id = b.user_id
		where p.user_id = $1`, userId)

	if err != nil {
		log.Fatal(err)
	}

	var positionAggregations []PositionAggregationModel
	var currentValue float32

	for aggregation.Next() {
		var symbol string
		var quantity float32
		var averagePrice float32
		var lastPrice float32
		var category int

		if err := aggregation.Scan(&symbol, &averagePrice, &quantity, &category, &lastPrice, &currentValue); err != nil {
			log.Println("Caindo nesse erro aqui 2")
			log.Fatal(err)
		} else {
			var asset AssetsModel = AssetsModel{
				Symbol:       symbol,
				Quantity:     quantity,
				AveragePrice: averagePrice,
				LastPrice:    lastPrice,
			}

			idx := sort.Search(len(positionAggregations), func(i int) bool {
				return positionAggregations[i].Category == category
			})

			if idx < len(positionAggregations) && positionAggregations[idx].Category == category {
				positionAggregations[idx].Assets = append(positionAggregations[idx].Assets, asset)
			} else {
				var aucAggregation PositionAggregationModel = PositionAggregationModel{
					Category:      category,
					TotalInvested: averagePrice * quantity,
					CurrentTotal:  lastPrice * quantity,
					Pnl:           lastPrice*quantity - averagePrice*quantity,
					PnlPercentage: ((lastPrice*quantity - averagePrice*quantity) / (averagePrice * quantity)) * 100,
					Assets:        []AssetsModel{asset},
				}

				positionAggregations = append(positionAggregations, aucAggregation)
			}
		}
	}

	aucAggregation := AucAggregationModel{
		TotalBalance:        currentValue,
		PositionAggregation: positionAggregations,
	}

	result, err := json.Marshal(aucAggregation)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Println("My result", string(result))

	fmt.Fprint(w, string(result))
}
