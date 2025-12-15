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

type RevokedTokenRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewRevokedTokenRepository(exec db.Executor, logger zerolog.Logger) *RevokedTokenRepository {
	return &RevokedTokenRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.revoked_token"),
	}
}

const (
	insertRevokedTokenSQL = `
		INSERT INTO revoked_tokens (jti, expires_at, orbit_id, reason, created_at)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, created_at
	`

	selectRevokedTokenByJTI = `
		SELECT id, jti, expires_at, orbit_id, reason, created_at
		FROM revoked_tokens
		WHERE jti = $1 AND orbit_id = $2
		LIMIT 1
	`
)

func (r *RevokedTokenRepository) Create(ctx context.Context, rt *models.RevokedToken) (*models.RevokedToken, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(
		ctx,
		insertRevokedTokenSQL,
		rt.JTI,
		rt.ExpiresAt,
		rt.OrbitID,
		rt.Reason,
		now,
	)

	if err := row.Scan(&rt.ID, &rt.CreatedAt); err != nil {
		r.logger.Error().Err(err).Str("jti", rt.JTI).Msg("revoked token insert failed")
		return nil, err
	}
	return rt, nil
}

func (r *RevokedTokenRepository) GetByJTI(ctx context.Context, orbitID int64, jti string) (*models.RevokedToken, error) {
	ctx, span := r.tracer.Start(ctx, "GetByJTI")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectRevokedTokenByJTI, jti, orbitID)
	rt := &models.RevokedToken{}
	if err := row.Scan(
		&rt.ID,
		&rt.JTI,
		&rt.ExpiresAt,
		&rt.OrbitID,
		&rt.Reason,
		&rt.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("jti", jti).Msg("revoked token get failed")
		return nil, err
	}
	return rt, nil
}
