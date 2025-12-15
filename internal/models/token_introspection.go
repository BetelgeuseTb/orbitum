package models

import "time"

type TokenIntrospection struct {
	ID        int64
	OrbitID   int64
	TokenJTI  string
	Active    bool
	Response  map[string]any
	ExpiresAt time.Time
	CreatedAt time.Time
}
