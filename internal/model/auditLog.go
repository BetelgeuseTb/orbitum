package model

import "time"

type AuditLog struct {
	ID        int64     `db:"id" json:"id"`
	UserID    *UUID     `db:"user_id" json:"user_id,omitempty"`
	Action    string    `db:"action" json:"action"`
	IPAddress *string   `db:"ip_address,omitempty" json:"ip_address,omitempty"`
	UserAgent *string   `db:"user_agent,omitempty" json:"user_agent,omitempty"`
	Metadata  []byte    `db:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
