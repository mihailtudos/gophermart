package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/repository"
	"github.com/mihailtudos/gophermart/internal/service/auth"
)

type userService struct {
	repo         repository.UserRepo
	tokenManager TokenManager
}

func NewUserService(repo repository.UserRepo,
	tm TokenManager) (*userService, error) {
	return &userService{
		repo:         repo,
		tokenManager: tm,
	}, nil
}

func (u *userService) Register(ctx context.Context, user domain.User) (int, error) {
	userID, err := u.repo.Create(ctx, user)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (u *userService) SetSessionToken(ctx context.Context, userID int, token string) error {
	st, err := u.tokenManager.CreateSession(userID, token)
	if err != nil {
		return err
	}

	return u.repo.SetSessionToken(ctx, st)
}

func (u *userService) Login(ctx context.Context, input auth.UserAuthInput) (domain.User, error) {
	user, err := u.repo.GetUserByLogin(ctx, input.Login)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (u *userService) RefreshTokens(ctx context.Context, refreshToken string) (auth.Tokens, error) {
	return auth.Tokens{}, nil
}

func (u *userService) GenerateUserTokens(ctx context.Context, userID int) (auth.Tokens, error) {
	uID := strconv.Itoa(userID)
	token, err := u.tokenManager.NewJWT(uID, nil)
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

func (u *userService) GetUserByLogin(ctx context.Context, login string) (domain.User, error) {
	return u.repo.GetUserByLogin(ctx, login)
}

func (u *userService) GetUserByID(ctx context.Context, userID int) (domain.User, error) {
	return u.repo.GetUserByID(ctx, userID)
}

func (u *userService) VerifyToken(ctx context.Context, token string) (int, error) {
	userID, err := u.tokenManager.Parse(token)
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user id %w", err)
	}

	return id, nil
}

func (u *userService) RegisterOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	return u.repo.RegisterOrder(ctx, order)
}

func (u *userService) GetUserOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	return u.repo.GetUserOrders(ctx, userID)
}

func (u *userService) GetUserBalance(ctx context.Context, userID int) (domain.UserBalance, error) {
	return u.repo.GetUserBalance(ctx, userID)
}

func (u *userService) WithdrawalPoints(ctx context.Context, wp domain.Withdrawal) (string, error) {
	return u.repo.WithdrawalPoints(ctx, wp)
}

func (u *userService) GetWithdrawals(ctx context.Context, userID int) ([]domain.Withdrawal, error) {
	return u.repo.GetWithdrawals(ctx, userID)
}

func (u *userService) GetUnfinishedOrders(ctx context.Context) ([]domain.Order, error) {
	return u.repo.GetUnfinishedOrders(ctx)
}

func (u *userService) UpdateOrder(ctx context.Context, order domain.Order) error {
	return u.repo.UpdateOrder(ctx, order)
}
