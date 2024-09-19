package domain

import "time"

type Order struct {
	ID         int       `json:"-"`
	UserID     int       `json:"user_id,omitempty"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual"`
	UploadedAt string    `json:"uploaded_at"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

const (
	ORDER_STATUS_NEW        = "NEW"
	ORDER_STATUS_PROCESSING = "PROCESSING"
	ORDER_STATUS_INVALID    = "INVALID"
	ORDER_STATUS_PROCESSED  = "PROCESSED"
)
