package di

import (
	"HubInvestments/home/application/service"
	persistence "HubInvestments/home/infra/persistency"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Container struct {
	UserService *service.AucService
}

func NewContainer() (*Container, error) {
	db, err := sqlx.Connect("postgres", "user=yanrodrigues dbname=yanrodrigues sslmode=disable password= host=localhost")

	if err != nil {
		return nil, err
	}

	// defer db.Close()

	// userRepo := persistence.NewSQLXUserRepository(db)
	// userRepo := persistence.NewSQLXUserRepository(db)
	userRepo := persistence.NewSQLXAucRepository(db)
	userService := service.NewUserService(userRepo)

	return &Container{
		UserService: userService,
	}, nil
}
