package di

import (
	balService "HubInvestments/balance/application/service"
	balancePersistence "HubInvestments/balance/infra/persistence"
	posService "HubInvestments/position/application/service"
	posUsecase "HubInvestments/position/application/usecase"
	aucPersistence "HubInvestments/position/infra/persistence"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Container interface {
	GetAucService() *posService.AucService
	GetBalanceService() *balService.BalanceService
	GetPositionAggregationUseCase() *posUsecase.GetPositionAggregationUseCase
}

type containerImpl struct {
	AucService                 *posService.AucService
	BalanceService             *balService.BalanceService
	PositionAggregationUseCase *posUsecase.GetPositionAggregationUseCase
}

func (c *containerImpl) GetAucService() *posService.AucService {
	return c.AucService
}

func (c *containerImpl) GetBalanceService() *balService.BalanceService {
	return c.BalanceService
}

func (c *containerImpl) GetPositionAggregationUseCase() *posUsecase.GetPositionAggregationUseCase {
	return c.PositionAggregationUseCase
}

func NewContainer() (Container, error) {
	db, err := sqlx.Connect("postgres", "user=yanrodrigues dbname=yanrodrigues sslmode=disable password= host=localhost")

	if err != nil {
		return nil, err
	}

	aucRepo := aucPersistence.NewSQLXAucRepository(db)
	aucService := posService.NewAucService(aucRepo)
	positionAggregationUseCase := posUsecase.NewGetPositionAggregationUseCase(aucRepo)

	balanceRepo := balancePersistence.NewSqlxBalanceRepository(db)
	balanceService := balService.NewBalanceService(balanceRepo)

	return &containerImpl{
		AucService:                 aucService,
		BalanceService:             balanceService,
		PositionAggregationUseCase: positionAggregationUseCase,
	}, nil
}
