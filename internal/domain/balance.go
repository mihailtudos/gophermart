package domain

import "time"

type UserBalance struct {
	ID        string    `json:"-"`
	UserID    string    `json:"-"`
	Current   float64   `json:"current"`
	Withdrawn float64   `json:"withdrawn"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
