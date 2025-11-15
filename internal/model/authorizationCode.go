package model

import "time"

type AuthorizationCode struct {
	Code                string    `db:"code" json:"code"`
	ClientID            UUID      `db:"client_id" json:"client_id"`
	UserID              UUID      `db:"user_id" json:"user_id"`
	Scopes              []string  `db:"scopes" json:"scopes"`
	RedirectURI         string    `db:"redirect_uri" json:"redirect_uri"`
	CodeChallenge       *string   `db:"code_challenge,omitempty" json:"code_challenge,omitempty"`
	CodeChallengeMethod *string   `db:"code_challenge_method,omitempty" json:"code_challenge_method,omitempty"`
	ExpiresAt           time.Time `db:"expires_at" json:"expires_at"`
	Used                bool      `db:"used" json:"used"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
}
