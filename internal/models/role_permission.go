package models

import "time"

type RolePermission struct {
	RoleID       int64
	PermissionID int64
	CreatedAt    time.Time
}
