package domain

import "time"

type Order struct {
	OrderNumber string    `json:"order" db:"order_number"`
	UserID      int       `json:"user_id,omitempty" db:"user_id"`
	OrderStatus string    `json:"status" db:"order_status"`
	Accrual     float64   `json:"accrual" db:"accrual"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

type UserOrder struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)
