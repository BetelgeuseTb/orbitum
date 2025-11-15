package sql

const (
	RoleInsert = `
        INSERT INTO orbitum.roles (name, description)
        VALUES ($1, $2)
        RETURNING id, name, description;
    `

	RoleGetByID = `
        SELECT id, name, description
        FROM orbitum.roles
        WHERE id = $1;
    `

	RoleGetByName = `
        SELECT id, name, description
        FROM orbitum.roles
        WHERE name = $1;
    `

	RoleList = `
        SELECT id, name, description
        FROM orbitum.roles
        ORDER BY id;
    `
)
