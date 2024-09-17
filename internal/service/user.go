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
	userId, err := u.repo.Create(ctx, user)

	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (u *userService) SetSessionToken(ctx context.Context, userID int, token string) error {
	st, err := u.tokenManager.CreateSession(userID, token)
	if err != nil {
		return err
	}

	return u.repo.SetSessionToken(ctx, st)
}

func (u *userService) Login(ctx context.Context, input auth.UserAuthInput) (domain.User, error) {
	user, err := u.repo.GetUserByEmail(ctx, input.Login)
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

func (u *userService) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	return u.repo.GetUserByEmail(ctx, email)
}

func (u *userService) GetUserByID(ctx context.Context, userID int) (domain.User, error) {
	return u.repo.GetUserById(ctx, userID)
}

func (u *userService) VerifyToken(ctx context.Context, token string) (int, error) {
	userId, err := u.tokenManager.Parse(token)
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(userId)
	if err != nil {
		return 0, fmt.Errorf("invalid user id %w", err)
	}

	return id, nil
}

func (u *userService) RegisterOrder(ctx context.Context, orderNumber string, userID int) (int, error) {
	order := domain.Order{
		Number: orderNumber,
		UserID: userID,
		Status: domain.NEW,
	}

	return u.repo.RegisterOrder(ctx, order)
}

func (u *userService) GetUserOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	return u.repo.GetUserOrders(ctx, userID)
}
