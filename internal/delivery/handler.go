package delivery

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mihailtudos/gophermart/internal/domain"
)

type AuthManager interface {
	SetSessionToken(ctx context.Context, userID string, tokens string) error
	GenerateUserTokens(ctx context.Context, userID string) (domain.Tokens, error)
	Login(ctx context.Context, input domain.UserAuthInput) (domain.User, error)
	Register(ctx context.Context, user domain.User) (string, error)
}
type UserManager interface {
	RegisterOrder(ctx context.Context, order domain.Order) (domain.Order, error)
	VerifyToken(ctx context.Context, token string) (string, error)
	GetUserByID(ctx context.Context, ID string) (domain.User, error)
	GetUserOrders(ctx context.Context, userID string) ([]domain.UserOrder, error)
	GetUnfinishedOrders(ctx context.Context) ([]domain.Order, error)
	UpdateOrder(ctx context.Context, updatedOrder domain.Order) error
	GetUserBalance(ctx context.Context, userID string) (domain.UserBalance, error)
	WithdrawalPoints(ctx context.Context, wp domain.Withdrawal) (string, error)
	GetWithdrawals(ctx context.Context, userID string) ([]domain.Withdrawal, error)
}

type Handler struct {
	Auth        AuthManager
	UserManager UserManager
}

func NewHandler(ah AuthManager, um UserManager) *chi.Mux {
	h := &Handler{
		Auth:        ah,
		UserManager: um,
	}

	router := chi.NewMux()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTION"},
		AllowedHeaders:   []string{"User-Agent", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	authHandler := NewAuthHanler(h.Auth)

	router.Route("/api/user", func(r chi.Router) {
		r.Mount("/", NewUserHandler(h.UserManager))
		r.Post("/login", authHandler.Signin)
		r.Post("/register", authHandler.Signup)
	})

	return router
}
