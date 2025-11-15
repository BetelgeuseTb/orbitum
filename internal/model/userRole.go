package model

import "time"

type UserRole struct {
	UserID     UUID      `db:"user_id" json:"user_id"`
	RoleID     int       `db:"role_id" json:"role_id"`
	AssignedAt time.Time `db:"assigned_at" json:"assigned_at"`
}
