package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mihailtudos/gophermart/internal/domain"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) (*userRepository, error) {
	return &userRepository{
		db: db,
	}, nil
}

func (u *userRepository) Create(ctx context.Context, user domain.User) (int, error) {
	stmt := `
		INSERT INTO users (login, password_hash)
		VALUES($1, $2)
		RETURNING id
	`

	row := u.db.QueryRowContext(ctx, stmt, user.Login, user.Password.Hash)
	if err := row.Err(); err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_login_key"`:
			return 0, ErrDuplicateLogin
		default:
			return 0, err
		}
	}

	var id int
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return int(id), nil
}

func (u *userRepository) SetSessionToken(ctx context.Context, st domain.Session) error {
	stmtDelete := `
		DELETE FROM session_tokens 
		WHERE user_id = $1
	`

	_, err := u.db.ExecContext(ctx, stmtDelete, st.UserID)
	if err != nil {
		return fmt.Errorf("error deleting old token: %w", err)
	}

	stmtInsert := `
		INSERT INTO session_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err = u.db.ExecContext(ctx, stmtInsert, st.UserID, st.Token, st.ExpiresAt)
	if err != nil {
		return fmt.Errorf("error inserting new token: %w", err)
	}

	return nil
}

func (u *userRepository) GetUserByLogin(ctx context.Context, login string) (domain.User, error) {
	stmt := `
		SELECT id, login, password_hash, created_at, version
		FROM users
		WHERE login = $1
	`
	row := u.db.QueryRowContext(ctx, stmt, login)
	if err := row.Err(); err != nil {
		return domain.User{}, err
	}

	user := domain.User{}

	err := row.Scan(
		&user.ID,
		&user.Login,
		&user.Password.Hash,
		&user.CreatedAt,
		&user.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return domain.User{}, ErrNoRowsFound
		default:
			return domain.User{}, nil
		}
	}

	return user, nil
}

func (u *userRepository) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	stmt := `
		SELECT id, login, version, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := domain.User{}

	row := u.db.QueryRowContext(ctx, stmt, id)
	if err := row.Err(); err != nil {
		return user, err
	}

	err := row.Scan(
		&user.ID,
		&user.Login,
		&user.Version,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (u *userRepository) RegisterOrder(ctx context.Context, order domain.Order) (int, error) {
	existingOrder := domain.Order{}

	checkOrderStmt := `
		SELECT id, user_id, status 
		FROM orders 
		WHERE number = $1
	`
	err := u.db.QueryRowContext(ctx, checkOrderStmt, order.Number).Scan(
		&existingOrder.ID,
		&existingOrder.UserID,
		&existingOrder.Status)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("error checking existing order: %w", err)
	}

	if err == nil {
		if existingOrder.UserID != order.UserID {
			return 0, ErrOrderAlreadyExistsDifferentUser
		}

		if existingOrder.UserID == order.UserID {
			if existingOrder.Status == domain.PROCESSING {
				return 0, ErrOrderAlreadyAccepted
			}

			return 0, ErrOrderAlreadyExistsSameUser
		}
	}

	insertStmt := `
		INSERT INTO orders(user_id, number, status)
		VALUES($1, $2, $3)
		RETURNING id
	`

	var newOrderID int
	err = u.db.QueryRowContext(ctx, insertStmt, order.UserID, order.Number, order.Status).Scan(&newOrderID)
	if err != nil {
		return 0, fmt.Errorf("error inserting new order: %w", err)
	}

	return newOrderID, nil
}

func (u *userRepository) GetUserOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	stmt := `
		SELECT number, created_at, status, accrual
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := u.db.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var createdAt time.Time
		order := domain.Order{}

		err := rows.Scan(
			&order.Number,
			&createdAt,
			&order.Status,
			&order.Accrual,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning order row: %w", err)
		}

		order.UploadedAt = createdAt.Format(time.RFC3339)

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after reading rows: %w", err)
	}

	return orders, nil
}
