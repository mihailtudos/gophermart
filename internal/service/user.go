package service

import (
	"context"

	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/repository"
	"github.com/mihailtudos/gophermart/internal/service/auth"
)

type UserService struct {
	repo         repository.UserRepo
	tokenManager TokenManager
}

func NewUserService(repo repository.UserRepo,
	tm TokenManager) (*UserService, error) {
	return &UserService{
		repo:         repo,
		tokenManager: tm,
	}, nil
}

func (u *UserService) Register(ctx context.Context, user domain.User) (string, error) {
	userID, err := u.repo.Create(ctx, user)

	if err != nil {
		return "", err
	}

	return userID, nil
}

func (u *UserService) SetSessionToken(ctx context.Context, userID string, token string) error {
	st, err := u.tokenManager.CreateSession(userID, token)
	if err != nil {
		return err
	}

	return u.repo.SetSessionToken(ctx, st)
}

func (u *UserService) Login(ctx context.Context, input auth.UserAuthInput) (domain.User, error) {
	user, err := u.repo.GetUserByLogin(ctx, input.Login)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (u *UserService) RefreshTokens(ctx context.Context, refreshToken string) (auth.Tokens, error) {
	return auth.Tokens{}, nil
}

func (u *UserService) GenerateUserTokens(ctx context.Context, userID string) (auth.Tokens, error) {
	token, err := u.tokenManager.NewJWT(userID, nil)
	if err != nil {
		return auth.Tokens{}, err
	}

	refToken, err := u.tokenManager.NewRefreshToken()
	if err != nil {
		return auth.Tokens{}, err
	}

	return auth.Tokens{
		AccessToken:  token,
		RefreshToken: refToken,
	}, nil
}

func (u *UserService) GetUserByLogin(ctx context.Context, login string) (domain.User, error) {
	return u.repo.GetUserByLogin(ctx, login)
}

func (u *UserService) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	return u.repo.GetUserByID(ctx, userID)
}

func (u *UserService) VerifyToken(ctx context.Context, token string) (string, error) {
	userID, err := u.tokenManager.Parse(token)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (u *UserService) RegisterOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	return u.repo.RegisterOrder(ctx, order)
}

func (u *UserService) GetUserOrders(ctx context.Context, userID string) ([]domain.UserOrder, error) {
	return u.repo.GetUserOrders(ctx, userID)
}

func (u *UserService) GetUserBalance(ctx context.Context, userID string) (domain.UserBalance, error) {
	return u.repo.GetUserBalance(ctx, userID)
}

func (u *UserService) WithdrawalPoints(ctx context.Context, wp domain.Withdrawal) (string, error) {
	return u.repo.WithdrawalPoints(ctx, wp)
}

func (u *UserService) GetWithdrawals(ctx context.Context, userID string) ([]domain.Withdrawal, error) {
	return u.repo.GetWithdrawals(ctx, userID)
}

func (u *UserService) GetUnfinishedOrders(ctx context.Context) ([]domain.Order, error) {
	return u.repo.GetUnfinishedOrders(ctx)
}

func (u *UserService) UpdateOrder(ctx context.Context, order domain.Order) error {
	return u.repo.UpdateOrder(ctx, order)
}
