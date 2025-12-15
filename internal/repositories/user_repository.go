package postgres

import (
	"context"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/db"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type UserRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewUserRepository(exec db.Executor, logger zerolog.Logger) *UserRepository {
	return &UserRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.user"),
	}
}

const (
	insertUserSQL = `
		INSERT INTO users (
			orbit_id, username, email, email_verified,
			password_hash, password_algo, last_password_change,
			display_name, profile, is_active, is_locked, mfa_enabled,
			metadata, created_at, updated_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
		)
		RETURNING id, created_at, updated_at, deleted_at
	`

	selectUserByIDSQL = `
		SELECT id, orbit_id, username, email, email_verified,
		       password_hash, password_algo, last_password_change,
		       display_name, profile, is_active, is_locked, mfa_enabled,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	selectUserByUsernameSQL = `
		SELECT id, orbit_id, username, email, email_verified,
		       password_hash, password_algo, last_password_change,
		       display_name, profile, is_active, is_locked, mfa_enabled,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE orbit_id = $1 AND username = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	selectUserByEmailSQL = `
		SELECT id, orbit_id, username, email, email_verified,
		       password_hash, password_algo, last_password_change,
		       display_name, profile, is_active, is_locked, mfa_enabled,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE orbit_id = $1 AND email = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	updateUserSQL = `
		UPDATE users
		SET username = $2,
		    email = $3,
		    email_verified = $4,
		    password_hash = $5,
		    password_algo = $6,
		    last_password_change = $7,
		    display_name = $8,
		    profile = $9,
		    is_active = $10,
		    is_locked = $11,
		    mfa_enabled = $12,
		    metadata = $13,
		    updated_at = $14
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at
	`

	softDeleteUserSQL = `
		UPDATE users
		SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`

	listUsersByOrbitSQL = `
		SELECT id, orbit_id, username, email, email_verified,
		       password_hash, password_algo, last_password_change,
		       display_name, profile, is_active, is_locked, mfa_enabled,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE orbit_id = $1 AND deleted_at IS NULL
		ORDER BY id
		LIMIT $2 OFFSET $3
	`
)

func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, insertUserSQL,
		user.OrbitID,
		user.Username,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.PasswordAlgo,
		user.LastPasswordChange,
		user.DisplayName,
		user.Profile,
		user.IsActive,
		user.IsLocked,
		user.MFAEnabled,
		user.Metadata,
		now,
		now,
	)

	var deletedAt *time.Time
	if err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &deletedAt); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			r.logger.Warn().Err(err).Str("username", user.Username).Str("email", user.Email).Msg("unique constraint violation on create user")
			return nil, err
		}
		r.logger.Error().Err(err).Str("username", user.Username).Msg("create user failed")
		return nil, err
	}
	user.DeletedAt = deletedAt
	r.logger.Debug().Int64("user_id", user.ID).Str("username", user.Username).Msg("user created")
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectUserByIDSQL, id)
	return scanUserRow(row)
}

func (r *UserRepository) GetByUsername(ctx context.Context, orbitID int64, username string) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "GetByUsername")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectUserByUsernameSQL, orbitID, username)
	return scanUserRow(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, orbitID int64, email string) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "GetByEmail")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectUserByEmailSQL, orbitID, email)
	return scanUserRow(row)
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, updateUserSQL,
		user.ID,
		user.Username,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.PasswordAlgo,
		user.LastPasswordChange,
		user.DisplayName,
		user.Profile,
		user.IsActive,
		user.IsLocked,
		user.MFAEnabled,
		user.Metadata,
		now,
	)

	var updatedAt time.Time
	if err := row.Scan(&updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			r.logger.Warn().Err(err).Int64("user_id", user.ID).Msg("unique constraint violation on update user")
			return nil, err
		}
		r.logger.Error().Err(err).Int64("user_id", user.ID).Msg("update user failed")
		return nil, err
	}
	user.UpdatedAt = updatedAt
	r.logger.Debug().Int64("user_id", user.ID).Msg("user updated")
	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	ctx, span := r.tracer.Start(ctx, "Delete")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, softDeleteUserSQL, id, now)
	var returnedID int64
	if err := row.Scan(&returnedID); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		r.logger.Error().Err(err).Int64("user_id", id).Msg("soft delete user failed")
		return err
	}
	r.logger.Info().Int64("user_id", id).Msg("user soft-deleted")
	return nil
}

func (r *UserRepository) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	rows, err := r.exec.Query(ctx, listUsersByOrbitSQL, orbitID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Int64("orbit_id", orbitID).Msg("list users query failed")
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u, err := scanUserRow(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func scanUserRow(scanner interface{ Scan(dest ...any) error }) (*models.User, error) {
	u := &models.User{}
	err := scanner.Scan(
		&u.ID,
		&u.OrbitID,
		&u.Username,
		&u.Email,
		&u.EmailVerified,
		&u.PasswordHash,
		&u.PasswordAlgo,
		&u.LastPasswordChange,
		&u.DisplayName,
		&u.Profile,
		&u.IsActive,
		&u.IsLocked,
		&u.MFAEnabled,
		&u.Metadata,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}
