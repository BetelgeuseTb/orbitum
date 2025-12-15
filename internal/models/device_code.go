package models

import "time"

type DeviceCode struct {
	ID              int64
	OrbitID         int64
	ClientID        int64
	DeviceCodeHash  string
	UserCode        string
	Scopes          []string
	ExpiresAt       time.Time
	PollIntervalSec int
	Status          DeviceCodeStatus
	UserID          *int64
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DeviceCodeStatus string

const (
	DeviceCodePending  DeviceCodeStatus = "pending"
	DeviceCodeApproved DeviceCodeStatus = "approved"
	DeviceCodeDenied   DeviceCodeStatus = "denied"
	DeviceCodeConsumed DeviceCodeStatus = "consumed"
	DeviceCodeExpired  DeviceCodeStatus = "expired"
)
