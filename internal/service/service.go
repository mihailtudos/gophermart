package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mihailtudos/gophermart/internal/config"
	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/logger"
	"github.com/mihailtudos/gophermart/internal/repository"
	"github.com/mihailtudos/gophermart/internal/service/auth"
)

type TokenManager interface {
	NewJWT(userID string, ttl *time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
	CreateSession(userID string, token string) (domain.Session, error)
}

type AccrualClient interface {
	GetOrderInfo(ctx context.Context, order domain.Order) (domain.Order, error)
}

// TODO - make use of interfaces
// type tokenManager interfect {
// ....
// }
// type userService interface {
// ....
// }

type Services struct {
	UserService   *UserService
	TokenManager  TokenManager
	AccrualClient AccrualClient
}

func NewServices(repos *repository.Repositories,
	authConfig config.AuthConfig,
	accrualClient AccrualClient) (*Services, error) {
	tms, err := auth.NewManager(authConfig.JWT)
	if err != nil {
		return nil, err
	}

	userService, err := NewUserService(repos.UserRepo, tms)
	if err != nil {
		return nil, err
	}

	return &Services{
		UserService:   userService,
		TokenManager:  tms,
		AccrualClient: accrualClient,
	}, nil
}

func (ss *Services) UpdateOrdersInBackground(ctx context.Context, jobInterval time.Duration) {
	ticker := time.NewTicker(jobInterval)

	go func() {
		defer func() {
			if p := recover(); p != nil {
				logger.Log.Warn("recover from panic ", slog.Any("err", p))
			}
		}()

		for {
			select {
			case <-ticker.C:
				ss.updateOrders(ctx)
			case <-ctx.Done():
				logger.Log.Info("sutting down the background process...")
				ticker.Stop()
				return
			}
		}
	}()
}

func (ss *Services) updateOrders(ctx context.Context) {
	orders, err := ss.UserService.GetUnfinishedOrders(ctx)
	if err != nil {
		logger.Log.Error("update orders: get unfinished orders", slog.String("err", err.Error()))
		return
	}

	if len(orders) == 0 {
		return
	}

	for _, v := range orders {
		fmt.Println(v.OrderNumber, v.OrderStatus)
	}

	for _, order := range orders {
		updateOrder, err := ss.AccrualClient.GetOrderInfo(ctx, order)
		if err != nil {
			logger.Log.Error("update orders: get order info", slog.String("err", err.Error()))
			continue
		}

		fmt.Printf("GetOrderInfo %#v", updateOrder)

		if order.OrderStatus != updateOrder.OrderStatus {
			err := ss.UserService.UpdateOrder(ctx, updateOrder)

			if err != nil {
				logger.Log.Error("failed to update the order status", slog.String("err", err.Error()))
			}

			continue
		}

		logger.Log.Info("skipping order " + order.OrderNumber)
	}
}
