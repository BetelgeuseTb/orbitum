package model

import "time"

type RevokedToken struct {
	JTI       UUID      `db:"jti" json:"jti"`
	TokenType string    `db:"token_type" json:"token_type"`
	RevokedAt time.Time `db:"revoked_at" json:"revoked_at"`
	Reason    *string   `db:"reason,omitempty" json:"reason,omitempty"`
}
