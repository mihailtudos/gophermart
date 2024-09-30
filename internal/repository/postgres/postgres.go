package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/mihailtudos/gophermart/internal/config"
)

func NewPostgresDB(ctx context.Context, cfg config.DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", cfg.DSN)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(10 * time.Second)
	db.SetMaxOpenConns(95)

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
