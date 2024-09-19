package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mihailtudos/gophermart/internal/app/accrual"
	"github.com/mihailtudos/gophermart/internal/config"
	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/logger"
	"github.com/mihailtudos/gophermart/internal/repository"
	"github.com/mihailtudos/gophermart/internal/service/auth"
)

type UserService interface {
	Register(ctx context.Context, input domain.User) (int, error)
	Login(ctx context.Context, input auth.UserAuthInput) (domain.User, error)
	GenerateUserTokens(ctx context.Context, userID int) (auth.Tokens, error)
	SetSessionToken(ctx context.Context, userID int, token string) error
	RefreshTokens(ctx context.Context, refreshToken string) (auth.Tokens, error)
	GetUserByLogin(ctx context.Context, login string) (domain.User, error)
	GetUserByID(ctx context.Context, userID int) (domain.User, error)
	VerifyToken(ctx context.Context, token string) (int, error)

	repository.OrdersHandler
	repository.UserBalance
	repository.BalanceHandler
}

type OrderService interface {
	Create(ctx context.Context, orderNumber string) (int, error)
}

type TokenManager interface {
	NewJWT(userID string, ttl *time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
	CreateSession(userID int, token string) (domain.Session, error)
}

type Services struct {
	UserService
	OrderService
	TokenManager
	accrual.AccrualClient
}

func NewServices(repos *repository.Repositories,
	authConfig config.AuthConfig, accrualClient accrual.AccrualClient) (*Services, error) {
	tms, err := auth.NewManager(authConfig.JWT)
	if err != nil {
		return nil, err
	}

	us, err := NewUserService(repos.UserRepo, tms)
	if err != nil {
		return nil, err
	}

	os, err := NewOrderService(repos.OrderRepo)
	if err != nil {
		return nil, err
	}

	return &Services{
		UserService:   us,
		OrderService:  os,
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
		fmt.Println(v.Number, v.Status)
	}

	for _, order := range orders {
		updateOrder, err := ss.AccrualClient.GetOrderInfo(order)
		if err != nil {
			logger.Log.Error("update orders: get order info", slog.String("err", err.Error()))
			continue
		}

		fmt.Printf("GetOrderInfo %#v", updateOrder)

		if order.Status != updateOrder.Status {
			err := ss.UserService.UpdateOrder(ctx, updateOrder)
			if err != nil {
				logger.Log.Error("failed to update the order status", slog.String("err", err.Error()))
				continue
			}
		} else {
			logger.Log.Info("skipping order " +  order.Number)
		}
	}
}
