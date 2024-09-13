package service

import (
	"context"
	"time"

	"github.com/mihailtudos/gophermart/internal/config"
	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/repository"
	"github.com/mihailtudos/gophermart/internal/service/auth"
)

type UserService interface {
	Register(ctx context.Context, input domain.User) (int, error)
	Login(ctx context.Context, input auth.UserAuthInput) (domain.User, error)
	GenerateUserTokens(ctx context.Context, userID int) (auth.Tokens, error)
	SetSessionToken(ctx context.Context, userID int, token string) error
	RefreshTokens(ctx context.Context, refreshToken string) (auth.Tokens, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByID(ctx context.Context, userID int) (domain.User, error)
	VerifyToken(ctx context.Context, token string) (int, error)
	RegisterOrder(ctx context.Context, orderNumber string, userId int) (int, error)
	GetUserOrders(ctx context.Context, userID int) ([]domain.Order, error)
}

type OrderService interface {
	Create(ctx context.Context, orderNumber string) (int, error)
}

type TokenManager interface {
	NewJWT(userId string, ttl *time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
	CreateSession(userId int, token string) (domain.Session, error)
}

type Services struct {
	UserService
	OrderService
	TokenManager
}

func NewServices(repos *repository.Repositories, authConfig config.AuthConfig) (*Services, error) {
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
		UserService:  us,
		OrderService: os,
		TokenManager: tms,
	}, nil
}
