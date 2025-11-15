package sql

const (
    UserInsert = `
        INSERT INTO orbitum.users (email, password_hash)
        VALUES ($1, $2)
        RETURNING id, email, password_hash, is_active, created_at, updated_at;
    `

    UserGetByID = `
        SELECT id, email, password_hash, is_active, created_at, updated_at
        FROM orbitum.users
        WHERE id = $1;
    `

    UserGetByEmail = `
        SELECT id, email, password_hash, is_active, created_at, updated_at
        FROM orbitum.users
        WHERE email = $1;
    `

    UserUpdatePassword = `
        UPDATE orbitum.users
        SET password_hash = $2, updated_at = NOW()
        WHERE id = $1
        RETURNING id, email, password_hash, is_active, created_at, updated_at;
    `

    UserSetActive = `
        UPDATE orbitum.users
        SET is_active = $2, updated_at = NOW()
        WHERE id = $1;
    `
)
