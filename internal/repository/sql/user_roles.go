package sql

const (
	UserRoleAssign = `
        INSERT INTO orbitum.user_roles (user_id, role_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING;
    `

	UserRoleRemove = `
        DELETE FROM orbitum.user_roles
        WHERE user_id = $1 AND role_id = $2;
    `

	UserRolesList = `
        SELECT r.id, r.name, r.description
        FROM orbitum.user_roles ur
        JOIN orbitum.roles r ON ur.role_id = r.id
        WHERE ur.user_id = $1;
    `
)
