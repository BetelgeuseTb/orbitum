package models

import "time"

type Permission struct {
	ID        int64
	OrbitID   int64
	Name      string
	Metadata  map[string]any
	CreatedAt time.Time
}
