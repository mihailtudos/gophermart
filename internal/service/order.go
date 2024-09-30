package service

import (
	"context"

	"github.com/mihailtudos/gophermart/internal/repository"
)

type orderService struct {
	repo repository.OrderRepo
}

func NewOrderService(orderRepo repository.OrderRepo) (*orderService, error) {
	return &orderService{
		repo: orderRepo,
	}, nil
}

func (os *orderService) Create(ctx context.Context, orderNumber string) (int, error) {
	return os.repo.Create(ctx, orderNumber)
}
