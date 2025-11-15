package model

import "time"

type AccessTokenRecord struct {
	TokenID  UUID      `db:"token_id" json:"token_id"`
	ClientID UUID      `db:"client_id" json:"client_id"`
	UserID   UUID      `db:"user_id,omitempty" json:"user_id,omitempty"`
	Scopes   []string  `db:"scopes" json:"scopes"`
	IssuedAt time.Time `db:"issued_at" json:"issued_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	JTI      UUID      `db:"jti" json:"jti"`
}
