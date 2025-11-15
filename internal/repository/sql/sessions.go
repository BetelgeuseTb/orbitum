package sql

const (
	SessionCreate = `
        INSERT INTO orbitum.sessions (user_id, refresh_token, user_agent, ip_address, expires_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at;
    `

	SessionGetByRefreshToken = `
        SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
        FROM orbitum.sessions
        WHERE refresh_token = $1;
    `

	SessionDelete = `
        DELETE FROM orbitum.sessions
        WHERE id = $1;
    `

	SessionDeleteByUser = `
        DELETE FROM orbitum.sessions
        WHERE user_id = $1;
    `
)
