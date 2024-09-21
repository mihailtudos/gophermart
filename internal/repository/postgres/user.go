package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/logger"
	"github.com/mihailtudos/gophermart/internal/repository/postgres/queries"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) (*userRepository, error) {
	return &userRepository{
		db: db,
	}, nil
}

// TODO - checkout https://sqlc.dev/
func (u *userRepository) Create(ctx context.Context, user domain.User) (int, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	userStmt := `
		INSERT INTO users (login, password_hash)
		VALUES($1, $2)
		RETURNING id
	`

	var userID int
	row := tx.QueryRowContext(ctx, userStmt, user.Login, user.Password.Hash)
	if err = row.Scan(&userID); err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_login_key"` {
			return 0, ErrDuplicateLogin
		}
		return 0, err
	}

	loyaltyStmt := `
		INSERT INTO user_loyalty_points (user_id)
		VALUES ($1)
	`

	_, err = tx.ExecContext(ctx, loyaltyStmt, userID)

	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return userID, nil
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

func (u *userRepository) RegisterOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return domain.Order{}, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	existingOrder := domain.Order{}
	err = tx.QueryRowContext(ctx, queries.GetOrderByOrderNumber, order.OrderNumber).Scan(
		&existingOrder.OrderNumber,
		&existingOrder.UserID,
		&existingOrder.OrderStatus)

	if err != nil && err != sql.ErrNoRows {
		return domain.Order{}, fmt.Errorf("error checking existing order: %w", err)
	}

	if err == nil {
		if existingOrder.UserID != order.UserID {
			return existingOrder, ErrOrderAlreadyExistsDifferentUser
		}

		if existingOrder.UserID == order.UserID {
			if existingOrder.OrderStatus == domain.OrderStatusProcessing {
				return existingOrder, ErrOrderAlreadyAccepted
			}

			return existingOrder, ErrOrderAlreadyExistsSameUser
		}
	}

	var insertedOrder domain.Order
	err = tx.QueryRowContext(ctx, queries.InsertOrderRecord, order.UserID, order.OrderNumber, order.OrderStatus).Scan(
		&insertedOrder.OrderNumber,
		&insertedOrder.UserID,
		&insertedOrder.OrderStatus,
		&insertedOrder.Accrual,
		&insertedOrder.CreatedAt,
		&insertedOrder.UpdatedAt,
	)

	if err != nil {
		return domain.Order{}, fmt.Errorf("error inserting new order: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.Order{}, err
	}

	return insertedOrder, nil
}

func (u *userRepository) GetUserOrders(ctx context.Context, userID int) ([]domain.UserOrder, error) {
	rows, err := u.db.QueryContext(ctx, queries.GetUserOrders, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []domain.UserOrder
	for rows.Next() {
		var createdAt time.Time
		order := domain.UserOrder{}

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

func (u *userRepository) GetUserBalance(ctx context.Context, userID int) (domain.UserBalance, error) {
	return u.getUserBalance(ctx, userID)
}

func (u *userRepository) WithdrawalPoints(ctx context.Context, wp domain.Withdrawal) (string, error) {
	var id string

	// Get the current user balance
	balance, err := u.getUserBalance(ctx, wp.UserID)
	if err != nil {
		return id, err
	}

	// Check if the user has sufficient points
	if balance.Current < wp.Sum {
		return id, ErrInsufficientPoints
	}

	// Begin a transaction
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return id, err
	}

	// Defer rollback in case of error, but avoid reusing the outer `err`
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Update the user's balance
	balance.Withdrawn += wp.Sum
	balance.Current -= wp.Sum

	// Update the user balance in the database
	_, err = tx.ExecContext(ctx, queries.UpdateUserBalance, balance.Current, balance.Withdrawn, wp.UserID)
	if err != nil {
		return id, err
	}

	// Create the withdrawal points record
	err = tx.QueryRowContext(ctx, queries.CreateWithdrawalPointsRecord, wp.UserID, wp.Order, wp.Sum).Scan(&id)
	if err != nil {
		return id, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return id, err
	}

	return id, nil
}

func (u *userRepository) getUserBalance(ctx context.Context, userID int) (domain.UserBalance, error) {
	var balance domain.UserBalance

	// Query the user's balance
	row := u.db.QueryRowContext(ctx, queries.GetUserBalanceStmt, userID)

	// Scan the result into the balance struct
	if err := row.Scan(&balance.Current, &balance.Withdrawn); err != nil {
		return balance, err
	}

	return balance, nil
}

func (u *userRepository) GetWithdrawals(ctx context.Context, userID int) ([]domain.Withdrawal, error) {
	var withdrawals []domain.Withdrawal

	// Replace 'queries.CreateWithdrawalPointsRecord' with the correct query for fetching withdrawals
	rows, err := u.db.QueryContext(ctx, queries.GetUserWithdrawals, userID)
	if err != nil {
		return withdrawals, err
	}
	defer rows.Close()

	// Iterate through the result set
	for rows.Next() {
		var withdrawal domain.Withdrawal
		var createdAt time.Time

		// Scan the row into the variables and check for errors
		if err := rows.Scan(
			&withdrawal.Order,
			&withdrawal.Sum,
			&createdAt,
		); err != nil {
			return withdrawals, err
		}

		// Format the timestamp and add the withdrawal to the slice
		withdrawal.ProcessedAt = createdAt.Format(time.RFC3339)
		withdrawals = append(withdrawals, withdrawal)
	}

	// Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		return withdrawals, err
	}

	return withdrawals, nil
}

func (u *userRepository) GetUnfinishedOrders(ctx context.Context) ([]domain.Order, error) {
	var orders []domain.Order

	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return orders, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	rows, err := tx.QueryContext(ctx, queries.GetUnfinishedOrders, domain.OrderStatusNew, domain.OrderStatusProcessed)

	if err != nil {
		return orders, err
	}

	defer rows.Close()

	for rows.Next() {
		var order domain.Order
		rows.Scan(
			&order.OrderNumber,
			&order.OrderStatus,
			&order.Accrual)

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return orders, err
	}

	if err = tx.Commit(); err != nil {
		return orders, err
	}

	return orders, nil
}

func (u *userRepository) UpdateOrder(ctx context.Context, order domain.Order) error {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = tx.QueryRowContext(ctx, queries.UpdateOrderStatusAndAccrualPoints,
		order.OrderStatus, order.Accrual, order.OrderNumber).Scan(
		&order.UserID,
	)

	if err != nil {
		return err
	}

	// TODO -- remove
	logger.Log.InfoContext(ctx,
		"status and accrual points updated",
		slog.String("order", order.OrderNumber),
		slog.Int("userID", order.UserID))

	res, err := tx.ExecContext(ctx, queries.UpdateUserLoyaltyPoints, order.Accrual, order.UserID)

	if err != nil {
		return err
	}

	ar, err := res.RowsAffected()
	if err != nil || ar != 1 {
		return fmt.Errorf("user loyalty points not updated %w", err)
	}

	logger.Log.InfoContext(ctx,
		"user loyalty points updated",
		slog.String("order", order.OrderNumber),
		slog.Int("userID", order.UserID))

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
