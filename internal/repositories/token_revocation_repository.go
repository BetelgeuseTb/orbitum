package postgres

import (
	"context"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/db"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type TokenRevocationRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewTokenRevocationRepository(exec db.Executor, logger zerolog.Logger) *TokenRevocationRepository {
	return &TokenRevocationRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.token_revocation"),
	}
}

const (
	insertTokenRevocationSQL = `
		INSERT INTO token_revocations (
			orbit_id, token_jti, token_type, reason, revoked_at, revoked_by, metadata
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id
	`
)

func (r *TokenRevocationRepository) Create(ctx context.Context, tr *models.TokenRevocation) (*models.TokenRevocation, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	tr.RevokedAt = time.Now().UTC()
	row := r.exec.QueryRow(
		ctx,
		insertTokenRevocationSQL,
		tr.OrbitID,
		tr.TokenJTI,
		tr.TokenType,
		tr.Reason,
		tr.RevokedAt,
		tr.RevokedBy,
		tr.Metadata,
	)

	if err := row.Scan(&tr.ID); err != nil {
		r.logger.Error().Err(err).Str("jti", tr.TokenJTI).Msg("token revocation insert failed")
		return nil, err
	}
	return tr, nil
}
