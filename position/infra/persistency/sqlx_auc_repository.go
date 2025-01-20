package persistence

import (
	domain "HubInvestments/position/domain/model"
	"HubInvestments/position/domain/repository"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SQLXAucRepository struct {
	db *sqlx.DB
}

func NewSQLXAucRepository(db *sqlx.DB) repository.AucRepository {
	return &SQLXAucRepository{db: db}
}

func (r *SQLXAucRepository) GetPositionAggregation(userId string) ([]domain.AssetsModel, error) {
	//TODO: testar r.db.Get. Parece que ja serializa uma struct direto
	//err := r.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	aggregation, err := r.db.Queryx(`
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
		println(err)
		return nil, fmt.Errorf(err.Error())
	}

	var positionAggregations []domain.AssetsModel

	for aggregation.Next() {
		var symbol string
		var quantity float32
		var averagePrice float32
		var lastPrice float32
		var category int
		var currentValue float32

		if err := aggregation.Scan(&symbol, &averagePrice, &quantity, &category, &lastPrice, &currentValue); err != nil {
			return nil, fmt.Errorf(err.Error())
		}

		var asset domain.AssetsModel = domain.AssetsModel{
			Symbol:       symbol,
			AveragePrice: averagePrice,
			Quantity:     quantity,
			Category:     category,
			LastPrice:    lastPrice,
		}

		positionAggregations = append(positionAggregations, asset)
	}

	return positionAggregations, nil
}
