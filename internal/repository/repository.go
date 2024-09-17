package repository

import (
	"context"
	"embed"

	"github.com/jmoiron/sqlx"
	"github.com/mihailtudos/gophermart/internal/config"
	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/logger"
	"github.com/mihailtudos/gophermart/internal/repository/postgres"
	"github.com/pressly/goose/v3"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

type UserRepo interface {
	Create(ctx context.Context, user domain.User) (int, error)
	GetUserByLogin(ctx context.Context, login string) (domain.User, error)
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	SetSessionToken(ctx context.Context, st domain.Session) error
	RegisterOrder(ctx context.Context, st domain.Order) (int, error)
	GetUserOrders(ctx context.Context, userID int) ([]domain.Order, error)
}

type OrderRepo interface {
	Create(ctx context.Context, number string) (int, error)
}

type Repositories struct {
	DB *sqlx.DB
	UserRepo
	OrderRepo
}

func NewRepository(ctx context.Context, dbConfig config.DBConfig) (*Repositories, error) {
	db, err := postgres.NewPostgresDB(ctx, dbConfig)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(ctx, db); err != nil {
		return nil, err
	}

	userRepo, err := postgres.NewUserRepository(db)
	if err != nil {
		return nil, err
	}

	orderRepo, err := postgres.NewOrderRepository(db)
	if err != nil {
		return nil, err
	}

	return &Repositories{
		UserRepo:  userRepo,
		OrderRepo: orderRepo,
		DB:        db,
	}, nil
}

func (r *Repositories) Close() error {
	if r.DB != nil {
		return r.DB.Close()
	}

	return nil
}

func runMigrations(ctx context.Context, db *sqlx.DB) error {
	goose.SetBaseFS(migrations)

	// Access the underlying *sql.DB from *sqlx.DB
	sqlDB := db.DB

	// Run migrations, specify the embedded path
	if err := goose.Up(sqlDB, "db/migrations"); err != nil {
		return err
	}

	logger.Log.InfoContext(ctx, "migrations applied successfully!")

	return nil
}
