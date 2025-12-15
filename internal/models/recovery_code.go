package models

import "time"

type RecoveryCode struct {
	ID        int64
	UserID    int64
	CodeHash  string
	UsedAt    *time.Time
	CreatedAt time.Time
}
