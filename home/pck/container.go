package di

import (
	"HubInvestments/home/application/service"
	persistence "HubInvestments/home/infra/persistency"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Container struct {
	AucService *service.AucService
}

func NewContainer() (*Container, error) {
	db, err := sqlx.Connect("postgres", "user=yanrodrigues dbname=yanrodrigues sslmode=disable password= host=localhost")

	if err != nil {
		return nil, err
	}

	userRepo := persistence.NewSQLXAucRepository(db)
	userService := service.NewUserService(userRepo)

	return &Container{
		AucService: userService,
	}, nil
}
