package sql

const (
	AuditInsert = `
        INSERT INTO orbitum.audit_logs (user_id, action, ip_address, user_agent)
        VALUES ($1, $2, $3, $4);
    `

	AuditListByUser = `
        SELECT id, user_id, action, ip_address, user_agent, created_at
        FROM orbitum.audit_logs
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3;
    `
)
