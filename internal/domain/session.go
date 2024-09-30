package domain

import "time"

type Session struct {
	ID         string    `json:"-"`
	UserID     string    `json:"-"`
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"-"`
	CreatedAt  time.Time `json:"-"`
	DeviceInfo *string   `json:"-"`
	IPAddress  *string   `json:"-"`
}
