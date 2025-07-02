package di

import (
	balService "HubInvestments/balance/application/service"
	balancePersistence "HubInvestments/balance/infra/persistence"
	posService "HubInvestments/position/application/service"
	aucPersistence "HubInvestments/position/infra/persistency"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Container interface {
	GetAucService() *posService.AucService
	GetBalanceService() *balService.BalanceService
}

type containerImpl struct {
	AucService     *posService.AucService
	BalanceService *balService.BalanceService
}

func (c *containerImpl) GetAucService() *posService.AucService {
	return c.AucService
}

func (c *containerImpl) GetBalanceService() *balService.BalanceService {
	return c.BalanceService
}

func NewContainer() (Container, error) {
	db, err := sqlx.Connect("postgres", "user=yanrodrigues dbname=yanrodrigues sslmode=disable password= host=localhost")

	if err != nil {
		return nil, err
	}

	aucRepo := aucPersistence.NewSQLXAucRepository(db)
	aucService := posService.NewAucService(aucRepo)

	balanceRepo := balancePersistence.NewSqlxBalanceRepository(db)
	balanceService := balService.NewBalanceService(balanceRepo)

	return &containerImpl{
		AucService:     aucService,
		BalanceService: balanceService,
	}, nil
}
