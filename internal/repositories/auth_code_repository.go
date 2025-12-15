package postgres

import (
	"context"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/db"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type AuthCodeRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewAuthCodeRepository(exec db.Executor, logger zerolog.Logger) *AuthCodeRepository {
	return &AuthCodeRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.auth_code"),
	}
}

const (
	insertAuthCodeSQL = `
		INSERT INTO auth_codes
			(code, orbit_id, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, used, metadata, created_at, expires_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, expires_at
	`

	selectAuthCodeByIDSQL = `
		SELECT id, code, orbit_id, client_id, user_id, redirect_uri, scope,
		       code_challenge, code_challenge_method, used, metadata, created_at, expires_at
		FROM auth_codes
		WHERE id = $1
		LIMIT 1
	`

	selectAuthCodeByCodeSQL = `
		SELECT id, code, orbit_id, client_id, user_id, redirect_uri, scope,
		       code_challenge, code_challenge_method, used, metadata, created_at, expires_at
		FROM auth_codes
		WHERE code = $1
		LIMIT 1
	`

	updateAuthCodeSetUsedByCodeSQL = `
		UPDATE auth_codes
		SET used = TRUE
		WHERE code = $1 AND used = FALSE
		RETURNING id
	`

	softDeleteAuthCodeSQL = `
		UPDATE auth_codes
		SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`

	updateAuthCodeSQL = `
		UPDATE auth_codes
		SET redirect_uri = $2, scope = $3, metadata = $4, expires_at = $5
		WHERE id = $1
		RETURNING id, expires_at
	`
)

func (r *AuthCodeRepository) Create(ctx context.Context, code *models.AuthCode) (*models.AuthCode, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	if code.CreatedAt.IsZero() {
		code.CreatedAt = now
	}
	row := r.exec.QueryRow(ctx, insertAuthCodeSQL,
		code.Code,
		code.OrbitID,
		code.ClientID,
		code.UserID,
		code.RedirectURI,
		code.Scope,
		code.CodeChallenge,
		code.CodeChallengeMethod,
		code.Used,
		code.Metadata,
		code.CreatedAt,
		code.ExpiresAt,
	)

	if err := row.Scan(&code.ID, &code.CreatedAt, &code.ExpiresAt); err != nil {
		r.logger.Error().Err(err).Str("code", code.Code).Msg("auth code create failed")
		return nil, err
	}

	r.logger.Debug().Int64("auth_code_id", code.ID).Str("code", code.Code).Msg("auth code created")
	return code, nil
}

func (r *AuthCodeRepository) GetByID(ctx context.Context, id int64) (*models.AuthCode, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectAuthCodeByIDSQL, id)
	ac, err := scanAuthCodeRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("auth_code_id", id).Msg("get auth code by id failed")
		return nil, err
	}
	return ac, nil
}

func (r *AuthCodeRepository) GetByCode(ctx context.Context, codeStr string) (*models.AuthCode, error) {
	ctx, span := r.tracer.Start(ctx, "GetByCode")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectAuthCodeByCodeSQL, codeStr)
	ac, err := scanAuthCodeRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("code", codeStr).Msg("get auth code by code failed")
		return nil, err
	}
	return ac, nil
}

func (r *AuthCodeRepository) SetUsedByCode(ctx context.Context, codeStr string) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "SetUsedByCode")
	defer span.End()

	row := r.exec.QueryRow(ctx, updateAuthCodeSetUsedByCodeSQL, codeStr)
	var id int64
	if err := row.Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		r.logger.Error().Err(err).Str("code", codeStr).Msg("set auth code used failed")
		return false, err
	}
	r.logger.Info().Int64("auth_code_id", id).Str("code", codeStr).Msg("auth code marked used")
	return true, nil
}

func (r *AuthCodeRepository) Update(ctx context.Context, code *models.AuthCode) (*models.AuthCode, error) {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	row := r.exec.QueryRow(ctx, updateAuthCodeSQL,
		code.ID,
		code.RedirectURI,
		code.Scope,
		code.Metadata,
		code.ExpiresAt,
	)

	var returnedID int64
	var returnedExpires time.Time
	if err := row.Scan(&returnedID, &returnedExpires); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("auth_code_id", code.ID).Msg("update auth code failed")
		return nil, err
	}
	code.ExpiresAt = returnedExpires
	code.ID = returnedID
	return code, nil
}

func (r *AuthCodeRepository) SoftDelete(ctx context.Context, id int64) error {
	ctx, span := r.tracer.Start(ctx, "SoftDelete")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, softDeleteAuthCodeSQL, id, now)
	var returnedID int64
	if err := row.Scan(&returnedID); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		r.logger.Error().Err(err).Int64("auth_code_id", id).Msg("soft delete auth code failed")
		return err
	}
	r.logger.Debug().Int64("auth_code_id", returnedID).Msg("auth code soft-deleted")
	return nil
}

func scanAuthCodeRow(scanner interface{ Scan(dest ...any) error }) (*models.AuthCode, error) {
	ac := &models.AuthCode{}
	err := scanner.Scan(
		&ac.ID,
		&ac.Code,
		&ac.OrbitID,
		&ac.ClientID,
		&ac.UserID,
		&ac.RedirectURI,
		&ac.Scope,
		&ac.CodeChallenge,
		&ac.CodeChallengeMethod,
		&ac.Used,
		&ac.Metadata,
		&ac.CreatedAt,
		&ac.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return ac, nil
}
