package domain

import "time"

type Withdrawal struct {
	ID          string    `json:"-"`
	UserID      int       `json:"-"`
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	ProcessedAt string    `json:"processed_at"`
}
