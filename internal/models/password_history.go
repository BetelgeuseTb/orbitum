package models

import "time"

type PasswordHistory struct {
	ID           int64
	UserID       int64
	PasswordHash string
	CreatedAt    time.Time
}
