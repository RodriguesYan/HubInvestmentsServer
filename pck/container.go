package di

import (
	balService "HubInvestments/balance/application/service"
	balUsecase "HubInvestments/balance/application/usecase"
	balancePersistence "HubInvestments/balance/infra/persistence"
	portfolioUsecase "HubInvestments/portfolio_summary/application/usecase"
	posService "HubInvestments/position/application/service"
	posUsecase "HubInvestments/position/application/usecase"
	positionPersistence "HubInvestments/position/infra/persistence"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Container interface {
	GetAucService() *posService.AucService
	GetPositionAggregationUseCase() *posUsecase.GetPositionAggregationUseCase
	GetBalanceService() *balService.BalanceService
	GetBalanceUseCase() *balUsecase.GetBalanceUseCase
	GetPortfolioSummaryUsecase() *portfolioUsecase.GetPortfolioSummaryUsecase
}

type containerImpl struct {
	AucService                 *posService.AucService
	BalanceService             *balService.BalanceService
	PositionAggregationUseCase *posUsecase.GetPositionAggregationUseCase
	BalanceUsecase             *balUsecase.GetBalanceUseCase
	PortfolioSummaryUsecase    *portfolioUsecase.GetPortfolioSummaryUsecase
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

func (c *containerImpl) GetBalanceUseCase() *balUsecase.GetBalanceUseCase {
	return c.BalanceUsecase
}

func (c *containerImpl) GetPortfolioSummaryUsecase() *portfolioUsecase.GetPortfolioSummaryUsecase {
	return c.PortfolioSummaryUsecase
}

func NewContainer() (Container, error) {
	db, err := sqlx.Connect("postgres", "user=yanrodrigues dbname=yanrodrigues sslmode=disable password= host=localhost")

	if err != nil {
		return nil, err
	}

	positionRepo := positionPersistence.NewSQLXPositionRepository(db)
	aucService := posService.NewAucService(positionRepo)
	positionAggregationUseCase := posUsecase.NewGetPositionAggregationUseCase(positionRepo)

	balanceRepo := balancePersistence.NewSqlxBalanceRepository(db)
	balanceService := balService.NewBalanceService(balanceRepo)
	balanceUsecase := balUsecase.NewGetBalanceUseCase(balanceService)

	portfolioSummaryUseCase := portfolioUsecase.NewGetPortfolioSummaryUsecase(*positionAggregationUseCase, *balanceUsecase)

	return &containerImpl{
		AucService:                 aucService,
		BalanceService:             balanceService,
		PositionAggregationUseCase: positionAggregationUseCase,
		BalanceUsecase:             balanceUsecase,
		PortfolioSummaryUsecase:    portfolioSummaryUseCase,
	}, nil
}
