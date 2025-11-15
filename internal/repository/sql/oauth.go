package sql

const (
	OAuthClientInsert = `
        INSERT INTO orbitum.oauth_clients
            (client_secret, client_name, redirect_uris, scopes, grant_types, token_endpoint_auth_method, public)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING client_id, client_secret, client_name, redirect_uris, scopes,
                  grant_types, token_endpoint_auth_method, public, created_at, updated_at;
    `

	OAuthClientGet = `
        SELECT client_id, client_secret, client_name, redirect_uris, scopes,
               grant_types, token_endpoint_auth_method, public, created_at, updated_at
        FROM orbitum.oauth_clients
        WHERE client_id = $1;
    `

	OAuthCodeInsert = `
        INSERT INTO orbitum.authorization_codes
            (code, client_id, user_id, scopes, redirect_uri, code_challenge, code_challenge_method, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
    `

	OAuthCodeGetValid = `
        SELECT code, client_id, user_id, scopes, redirect_uri, code_challenge,
               code_challenge_method, expires_at, used, created_at
        FROM orbitum.authorization_codes
        WHERE code = $1 AND used = FALSE AND expires_at > NOW();
    `

	OAuthCodeMarkUsed = `
        UPDATE orbitum.authorization_codes
        SET used = TRUE
        WHERE code = $1;
    `

	OAuthAccessTokenInsert = `
        INSERT INTO orbitum.access_tokens
            (client_id, user_id, scopes, expires_at)
        VALUES ($1, $2, $3, $4)
        RETURNING token_id, client_id, user_id, scopes, issued_at, expires_at, jti;
    `

	OAuthAccessTokenGet = `
        SELECT token_id, client_id, user_id, scopes, issued_at, expires_at, jti
        FROM orbitum.access_tokens
        WHERE token_id = $1;
    `
)
