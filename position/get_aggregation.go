package get_aggregation

import (
	di "HubInvestments/pck"
	domain "HubInvestments/position/domain/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
)

// TODO: makefile dropando database, recriando tabelas e populando
// TODO: por docker pra banco de dados
// TODO: microsserviços (not requered)
// TODO: swagger
// TODO: rmq pra enfileirar envios de ordens
// TODO: GRPC
// TODO: websocket para cotaçao de ativos
//TODO: fazer testes pra login
//TODO: refatorar login em metodos menores

type TokenVerifier func(string, http.ResponseWriter) (string, error)

func GetAucAggregation(w http.ResponseWriter, r *http.Request, verifyToken TokenVerifier, container di.Container) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")

	userId, err := verifyToken(tokenString, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	assets, err := container.GetAucService().GetAucAggregation(userId)

	if err != nil {
		log.Fatalf("could not create user: %v", err)
	}

	var positionAggregations []domain.PositionAggregationModel
	var currentValue float32

	for _, element := range assets {
		index := sort.Search(len(positionAggregations), func(i int) bool {
			return positionAggregations[i].Category == element.Category
		})

		if index < len(positionAggregations) && positionAggregations[index].Category == element.Category {
			positionAggregations[index].Assets = append(positionAggregations[index].Assets, element)
			positionAggregations[index].TotalInvested += element.AveragePrice * element.Quantity
			positionAggregations[index].CurrentTotal += element.LastPrice * element.Quantity
			positionAggregations[index].Pnl += element.LastPrice*element.Quantity - element.AveragePrice*element.Quantity
			positionAggregations[index].PnlPercentage = ((element.LastPrice*element.Quantity - element.AveragePrice*element.Quantity) / (element.AveragePrice * element.Quantity)) * 100
		} else {
			var aucAggregation domain.PositionAggregationModel = domain.PositionAggregationModel{
				Category:      element.Category,
				TotalInvested: element.AveragePrice * element.Quantity,
				CurrentTotal:  element.LastPrice * element.Quantity,
				Pnl:           element.LastPrice*element.Quantity - element.AveragePrice*element.Quantity,
				PnlPercentage: ((element.LastPrice*element.Quantity - element.AveragePrice*element.Quantity) / (element.AveragePrice * element.Quantity)) * 100,
				Assets:        []domain.AssetsModel{element},
			}

			positionAggregations = append(positionAggregations, aucAggregation)
		}
	}

	aucAggregation := domain.AucAggregationModel{
		TotalBalance:        currentValue,
		PositionAggregation: positionAggregations,
	}

	result, err := json.Marshal(aucAggregation)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(result))
}
