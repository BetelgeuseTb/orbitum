package models

import "time"

type TokenRevocation struct {
	ID        int64
	OrbitID   int64
	TokenJTI  string
	TokenType string
	Reason    string
	RevokedAt time.Time
	RevokedBy *int64
	Metadata  map[string]any
}
