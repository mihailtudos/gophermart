package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/mihailtudos/gophermart/internal/config"
	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/repository/postgres"
)

type UserRepo interface {
	Create(ctx context.Context, user domain.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	SetSessionToken(ctx context.Context, st domain.Session) error
	RegisterOrder(ctx context.Context, st domain.Order) (int, error)
	GetUserOrders(ctx context.Context, userID int) ([]domain.Order, error)
}

type OrderRepo interface {
	Create(ctx context.Context, number string) (int, error)
}

type Repositories struct {
	DB *sqlx.DB
	UserRepo
	OrderRepo
}

func NewRepository(ctx context.Context, dbConfig config.DBConfig) (*Repositories, error) {
	db, err := postgres.NewPostgresDB(ctx, dbConfig)
	if err != nil {
		return nil, err
	}

	userRepo, err := postgres.NewUserRepository(db)
	if err != nil {
		return nil, err
	}

	orderRepo, err := postgres.NewOrderRepository(db)
	if err != nil {
		return nil, err
	}

	return &Repositories{
		UserRepo:  userRepo,
		OrderRepo: orderRepo,
		DB:        db,
	}, nil
}

func (r *Repositories) Close() error {
	if r.DB != nil {
		return r.DB.Close()
	}

	return nil
}
