package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/mihailtudos/gophermart/internal/config"
)

func NewPostgresDB(ctx context.Context, cfg config.DBConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s connect_timeout=3",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.SSLMode, cfg.Password,
	)

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("database connection timeout: %w", err)
		}
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("database ping timeout: %w", err)
		}
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}
