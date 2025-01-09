package di

import (
	"HubInvestments/home/application/service"
	persistence "HubInvestments/home/infra/persistency"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Container interface {
	GetAucService() *service.AucService
}

type containerImpl struct {
	AucService *service.AucService
}

func (c *containerImpl) GetAucService() *service.AucService {
	return c.AucService
}

func NewContainer() (Container, error) {
	db, err := sqlx.Connect("postgres", "user=yanrodrigues dbname=yanrodrigues sslmode=disable password= host=localhost")

	if err != nil {
		return nil, err
	}

	userRepo := persistence.NewSQLXAucRepository(db)
	userService := service.NewUserService(userRepo)

	return &containerImpl{
		AucService: userService,
	}, nil
}
