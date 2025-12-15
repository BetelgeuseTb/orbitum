package models

import "time"

type SecurityEvent struct {
	ID        int64
	OrbitID   int64
	UserID    *int64
	EventType string
	Severity  string
	Metadata  map[string]any
	CreatedAt time.Time
}
