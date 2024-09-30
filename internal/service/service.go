package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/logger"
)

type TokenManager interface {
	NewJWT(userID string, ttl *time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
	CreateSession(userID string, token string) (domain.Session, error)
}

type Auth interface {
	GenerateUserTokens(ctx context.Context, userID string) (domain.Tokens, error)
	SetSessionToken(ctx context.Context, userID string, tokens string) error
	Login(ctx context.Context, input domain.UserAuthInput) (domain.User, error)
	Register(ctx context.Context, user domain.User) (string, error)
	VerifyToken(ctx context.Context, token string) (string, error)
}

type AccrualClient interface {
	GetOrderInfo(ctx context.Context, order domain.Order) (domain.Order, error)
}

type UserManager interface {
	GetUserByID(ctx context.Context, ID string) (domain.User, error)
	GetUserBalance(ctx context.Context, userID string) (domain.UserBalance, error)
	WithdrawalPoints(ctx context.Context, wp domain.Withdrawal) (string, error)
	GetWithdrawals(ctx context.Context, userID string) ([]domain.Withdrawal, error)
	OrderService
	Auth
}

type OrderService interface {
	UpdateOrder(ctx context.Context, updateOrder domain.Order) error
	RegisterOrder(ctx context.Context, order domain.Order) (domain.Order, error)
	GetUserOrders(ctx context.Context, userID string) ([]domain.UserOrder, error)
	GetUnfinishedOrders(ctx context.Context) ([]domain.Order, error)
}

type Services struct {
	UserService   UserManager
	TokenManager  TokenManager
	AccrualClient AccrualClient
}

func NewServices(userService UserManager,
	tokenService TokenManager,
	accrualClient AccrualClient) (*Services, error) {

	return &Services{
		UserService:   userService,
		TokenManager:  tokenService,
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
