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

type AccessTokenRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewAccessTokenRepository(exec db.Executor, logger zerolog.Logger) *AccessTokenRepository {
	return &AccessTokenRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.access_token"),
	}
}

const (
	insertAccessTokenSQL = `
		INSERT INTO access_tokens
			(jti, orbit_id, client_id, user_id, is_jwt, token_string, scope,
			 issued_at, token_type, revoked, metadata, refresh_token_id, created_at, expires_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, created_at, expires_at
	`

	selectAccessTokenByIDSQL = `
		SELECT id, jti, orbit_id, client_id, user_id, is_jwt, token_string, scope,
		       issued_at, token_type, revoked, metadata, refresh_token_id, created_at, expires_at
		FROM access_tokens
		WHERE id = $1
		LIMIT 1
	`

	selectAccessTokenByJTISQL = `
		SELECT id, jti, orbit_id, client_id, user_id, is_jwt, token_string, scope,
		       issued_at, token_type, revoked, metadata, refresh_token_id, created_at, expires_at
		FROM access_tokens
		WHERE jti = $1
		LIMIT 1
	`

	updateAccessTokenSQL = `
		UPDATE access_tokens
		SET
			token_string = $2,
			revoked = $3,
			metadata = $4,
			refresh_token_id = $5,
			issued_at = $6,
			expires_at = $7
		WHERE id = $1
		RETURNING id, jti, orbit_id, client_id, user_id, is_jwt, token_string, scope,
		         issued_at, token_type, revoked, metadata, refresh_token_id, created_at, expires_at
	`

	revokeAccessTokenByJTISQL = `
		UPDATE access_tokens
		SET revoked = TRUE
		WHERE jti = $1 AND (revoked IS NULL OR revoked = FALSE)
		RETURNING id, revoked
	`
)

func (r *AccessTokenRepository) Create(ctx context.Context, token *models.AccessToken) (*models.AccessToken, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	row := r.exec.QueryRow(ctx, insertAccessTokenSQL,
		token.JTI,
		token.OrbitID,
		token.ClientID,
		token.UserID,
		token.IsJWT,
		token.TokenString,
		token.Scope,
		token.IssuedAt,
		token.TokenType,
		token.Revoked,
		token.Metadata,
		token.RefreshTokenID,
		token.CreatedAt,
		token.ExpiresAt,
	)

	if err := row.Scan(&token.ID, &token.CreatedAt, &token.ExpiresAt); err != nil {
		r.logger.Error().Err(err).Str("jti", token.JTI).Msg("access token create failed")
		return nil, err
	}

	r.logger.Debug().Int64("access_token_id", token.ID).Str("jti", token.JTI).Msg("access token created")
	return token, nil
}

func (r *AccessTokenRepository) GetByID(ctx context.Context, id int64) (*models.AccessToken, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectAccessTokenByIDSQL, id)
	at, err := scanAccessTokenRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("access_token_id", id).Msg("get access token by id failed")
		return nil, err
	}
	return at, nil
}

func (r *AccessTokenRepository) GetByJTI(ctx context.Context, jti string) (*models.AccessToken, error) {
	ctx, span := r.tracer.Start(ctx, "GetByJTI")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectAccessTokenByJTISQL, jti)
	at, err := scanAccessTokenRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("jti", jti).Msg("get access token by jti failed")
		return nil, err
	}
	return at, nil
}

func (r *AccessTokenRepository) Update(ctx context.Context, token *models.AccessToken) (*models.AccessToken, error) {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	row := r.exec.QueryRow(ctx, updateAccessTokenSQL,
		token.ID,
		token.TokenString,
		token.Revoked,
		token.Metadata,
		token.RefreshTokenID,
		token.IssuedAt,
		token.ExpiresAt,
	)

	var updated models.AccessToken
	if err := row.Scan(
		&updated.ID,
		&updated.JTI,
		&updated.OrbitID,
		&updated.ClientID,
		&updated.UserID,
		&updated.IsJWT,
		&updated.TokenString,
		&updated.Scope,
		&updated.IssuedAt,
		&updated.TokenType,
		&updated.Revoked,
		&updated.Metadata,
		&updated.RefreshTokenID,
		&updated.CreatedAt,
		&updated.ExpiresAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("access_token_id", token.ID).Msg("update access token failed")
		return nil, err
	}

	r.logger.Debug().Int64("access_token_id", updated.ID).Str("jti", updated.JTI).Msg("access token updated")
	return &updated, nil
}

func (r *AccessTokenRepository) RevokeByJTI(ctx context.Context, jti string) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "RevokeByJTI")
	defer span.End()

	row := r.exec.QueryRow(ctx, revokeAccessTokenByJTISQL, jti)
	var id int64
	var revoked bool
	if err := row.Scan(&id, &revoked); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		r.logger.Error().Err(err).Str("jti", jti).Msg("revoke access token failed")
		return false, err
	}

	r.logger.Info().Int64("access_token_id", id).Str("jti", jti).Bool("revoked", revoked).Msg("access token revoked")
	return revoked, nil
}

func scanAccessTokenRow(scanner interface{ Scan(dest ...any) error }) (*models.AccessToken, error) {
	at := &models.AccessToken{}
	err := scanner.Scan(
		&at.ID,
		&at.JTI,
		&at.OrbitID,
		&at.ClientID,
		&at.UserID,
		&at.IsJWT,
		&at.TokenString,
		&at.Scope,
		&at.IssuedAt,
		&at.TokenType,
		&at.Revoked,
		&at.Metadata,
		&at.RefreshTokenID,
		&at.CreatedAt,
		&at.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return at, nil
}
