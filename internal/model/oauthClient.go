package model

import "time"

type OAuthClient struct {
	ClientID                UUID      `db:"client_id" json:"client_id"`
	ClientSecretHash        string    `db:"client_secret_hash" json:"-"`
	ClientName              string    `db:"client_name" json:"client_name"`
	RedirectURIs            []string  `db:"redirect_uris" json:"redirect_uris"`
	Scopes                  []string  `db:"scopes" json:"scopes"`
	GrantTypes              []string  `db:"grant_types" json:"grant_types"`
	TokenEndpointAuthMethod string    `db:"token_endpoint_auth_method" json:"token_endpoint_auth_method"`
	Public                  bool      `db:"public" json:"public"`
	CreatedAt               time.Time `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time `db:"updated_at" json:"updated_at"`
}
