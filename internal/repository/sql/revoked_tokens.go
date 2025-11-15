package sql

const (
	TokenRevoke = `
        INSERT INTO orbitum.revoked_tokens (token_id)
        VALUES ($1)
        ON CONFLICT DO NOTHING;
    `

	TokenIsRevoked = `
        SELECT 1
        FROM orbitum.revoked_tokens
        WHERE token_id = $1;
    `
)
