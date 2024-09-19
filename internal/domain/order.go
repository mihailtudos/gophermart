package domain

import "time"

type Order struct {
	OrderNumber string    `json:"order" db:"order_number"`
	UserID      int       `json:"user_id,omitempty" db:"user_id"`
	OrderStatus string    `json:"status" db:"order_status"`
	Accrual     float64   `json:"accrual" db:"accrual"`
	UploadedAt  string    `json:"uploaded_at" db:"uploaded_at"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

const (
	ORDER_STATUS_NEW        = "NEW"
	ORDER_STATUS_PROCESSING = "PROCESSING"
	ORDER_STATUS_INVALID    = "INVALID"
	ORDER_STATUS_PROCESSED  = "PROCESSED"
)
