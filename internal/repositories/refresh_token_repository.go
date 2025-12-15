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

type RefreshTokenRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewRefreshTokenRepository(exec db.Executor, logger zerolog.Logger) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.refresh_token"),
	}
}

const (
	insertRefreshTokenSQL = `
		INSERT INTO refresh_tokens
			(expires_at, token_string, jti, orbit_id, client_id, user_id, revoked, rotated_from_id, rotated_to_id, scopes, metadata, last_used_at, use_count, created_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, created_at, expires_at
	`

	selectRefreshTokenByIDSQL = `
		SELECT id, expires_at, token_string, jti, orbit_id, client_id, user_id, revoked, rotated_from_id, rotated_to_id, scopes, metadata, last_used_at, use_count, created_at
		FROM refresh_tokens
		WHERE id = $1
		LIMIT 1
	`

	selectRefreshTokenByJTISQL = `
		SELECT id, expires_at, token_string, jti, orbit_id, client_id, user_id, revoked, rotated_from_id, rotated_to_id, scopes, metadata, last_used_at, use_count, created_at
		FROM refresh_tokens
		WHERE jti = $1
		LIMIT 1
	`

	updateRefreshTokenSQL = `
		UPDATE refresh_tokens
		SET
			token_string = $2,
			revoked = $3,
			rotated_from_id = $4,
			rotated_to_id = $5,
			scopes = $6,
			metadata = $7,
			last_used_at = $8,
			use_count = $9,
			expires_at = $10
		WHERE id = $1
		RETURNING id, expires_at
	`

	revokeRefreshTokenByJTISQL = `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE jti = $1 AND (revoked IS NULL OR revoked = FALSE)
		RETURNING id
	`

	rotateRefreshTokenSQL = `
		UPDATE refresh_tokens
		SET rotated_to_id = $2
		WHERE id = $1
		RETURNING id
	`
)

func (r *RefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) (*models.RefreshToken, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	row := r.exec.QueryRow(ctx, insertRefreshTokenSQL,
		token.ExpiresAt,
		token.TokenString,
		token.JTI,
		token.OrbitID,
		token.ClientID,
		token.UserID,
		token.Revoked,
		token.RotatedFromID,
		token.RotatedToID,
		token.Scopes,
		token.Metadata,
		token.LastUsedAt,
		token.UseCount,
		token.CreatedAt,
	)

	if err := row.Scan(&token.ID, &token.CreatedAt, &token.ExpiresAt); err != nil {
		r.logger.Error().Err(err).Str("jti", token.JTI).Msg("refresh token create failed")
		return nil, err
	}

	r.logger.Debug().Int64("refresh_token_id", token.ID).Str("jti", token.JTI).Msg("refresh token created")
	return token, nil
}

func (r *RefreshTokenRepository) GetByID(ctx context.Context, id int64) (*models.RefreshToken, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectRefreshTokenByIDSQL, id)
	rt, err := scanRefreshTokenRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("refresh_token_id", id).Msg("get refresh token by id failed")
		return nil, err
	}
	return rt, nil
}

func (r *RefreshTokenRepository) GetByJTI(ctx context.Context, jti string) (*models.RefreshToken, error) {
	ctx, span := r.tracer.Start(ctx, "GetByJTI")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectRefreshTokenByJTISQL, jti)
	rt, err := scanRefreshTokenRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("jti", jti).Msg("get refresh token by jti failed")
		return nil, err
	}
	return rt, nil
}

func (r *RefreshTokenRepository) Update(ctx context.Context, token *models.RefreshToken) (*models.RefreshToken, error) {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	row := r.exec.QueryRow(ctx, updateRefreshTokenSQL,
		token.ID,
		token.TokenString,
		token.Revoked,
		token.RotatedFromID,
		token.RotatedToID,
		token.Scopes,
		token.Metadata,
		token.LastUsedAt,
		token.UseCount,
		token.ExpiresAt,
	)

	var returnedID int64
	var returnedExpires time.Time
	if err := row.Scan(&returnedID, &returnedExpires); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("refresh_token_id", token.ID).Msg("update refresh token failed")
		return nil, err
	}
	token.ID = returnedID
	token.ExpiresAt = returnedExpires

	r.logger.Debug().Int64("refresh_token_id", token.ID).Str("jti", token.JTI).Msg("refresh token updated")
	return token, nil
}

func (r *RefreshTokenRepository) RevokeByJTI(ctx context.Context, jti string) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "RevokeByJTI")
	defer span.End()

	row := r.exec.QueryRow(ctx, revokeRefreshTokenByJTISQL, jti)
	var id int64
	if err := row.Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		r.logger.Error().Err(err).Str("jti", jti).Msg("revoke refresh token failed")
		return false, err
	}
	r.logger.Info().Int64("refresh_token_id", id).Str("jti", jti).Msg("refresh token revoked")
	return true, nil
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, id int64, rotatedToID int64) error {
	ctx, span := r.tracer.Start(ctx, "Rotate")
	defer span.End()

	row := r.exec.QueryRow(ctx, rotateRefreshTokenSQL, id, rotatedToID)
	var returnedID int64
	if err := row.Scan(&returnedID); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		r.logger.Error().Err(err).Int64("refresh_token_id", id).Int64("rotated_to_id", rotatedToID).Msg("rotate refresh token failed")
		return err
	}
	r.logger.Debug().Int64("refresh_token_id", returnedID).Int64("rotated_to_id", rotatedToID).Msg("refresh token rotated")
	return nil
}

func scanRefreshTokenRow(scanner interface{ Scan(dest ...any) error }) (*models.RefreshToken, error) {
	rt := &models.RefreshToken{}
	err := scanner.Scan(
		&rt.ID,
		&rt.ExpiresAt,
		&rt.TokenString,
		&rt.JTI,
		&rt.OrbitID,
		&rt.ClientID,
		&rt.UserID,
		&rt.Revoked,
		&rt.RotatedFromID,
		&rt.RotatedToID,
		&rt.Scopes,
		&rt.Metadata,
		&rt.LastUsedAt,
		&rt.UseCount,
		&rt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return rt, nil
}
