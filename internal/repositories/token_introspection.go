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

type TokenIntrospectionRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewTokenIntrospectionRepository(exec db.Executor, logger zerolog.Logger) *TokenIntrospectionRepository {
	return &TokenIntrospectionRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.token_introspection"),
	}
}

const (
	insertTokenIntrospectionSQL = `
		INSERT INTO token_introspections (
			orbit_id, token_jti, active, response, expires_at, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, created_at
	`

	selectTokenIntrospectionByJTI = `
		SELECT id, orbit_id, token_jti, active, response, expires_at, created_at
		FROM token_introspections
		WHERE token_jti = $1 AND orbit_id = $2
		LIMIT 1
	`
)

func (r *TokenIntrospectionRepository) Create(ctx context.Context, ti *models.TokenIntrospection) (*models.TokenIntrospection, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(
		ctx,
		insertTokenIntrospectionSQL,
		ti.OrbitID,
		ti.TokenJTI,
		ti.Active,
		ti.Response,
		ti.ExpiresAt,
		now,
	)

	if err := row.Scan(&ti.ID, &ti.CreatedAt); err != nil {
		r.logger.Error().Err(err).Str("jti", ti.TokenJTI).Msg("token introspection insert failed")
		return nil, err
	}
	return ti, nil
}

func (r *TokenIntrospectionRepository) GetByJTI(ctx context.Context, orbitID int64, jti string) (*models.TokenIntrospection, error) {
	ctx, span := r.tracer.Start(ctx, "GetByJTI")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectTokenIntrospectionByJTI, jti, orbitID)
	ti := &models.TokenIntrospection{}
	if err := row.Scan(
		&ti.ID,
		&ti.OrbitID,
		&ti.TokenJTI,
		&ti.Active,
		&ti.Response,
		&ti.ExpiresAt,
		&ti.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("jti", jti).Msg("token introspection get failed")
		return nil, err
	}
	return ti, nil
}
