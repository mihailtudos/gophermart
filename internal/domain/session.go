package domain

import "time"

// TODO - swap to UUID
type Session struct {
	ID         int       `json:"-"`
	UserID     int       `json:"-"`
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"-"`
	CreatedAt  time.Time `json:"-"`
	DeviceInfo *string   `json:"-"`
	IPAddress  *string   `json:"-"`
}
