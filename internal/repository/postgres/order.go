package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type orderRepository struct {
	DB *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) (*orderRepository, error) {
	return &orderRepository{
		DB: db,
	}, nil
}

func (o *orderRepository) Create(ctx context.Context, orderNumber string) (int, error) {
	return 0, nil
}
