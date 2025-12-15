package postgres

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/db"
	"github.com/jackc/pgx/v5"
)

type ConsentRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewConsentRepository(exec db.Executor, logger zerolog.Logger) *ConsentRepository {
	return &ConsentRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.consent"),
	}
}

const (
	insertConsentSQL = `
		INSERT INTO consents (
			orbit_id, user_id, client_id, scopes, granted_at,
			expires_at, revoked, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at, updated_at
	`

	selectConsentSQL = `
		SELECT id, orbit_id, user_id, client_id, scopes, granted_at,
		       expires_at, revoked, created_at, updated_at
		FROM consents
		WHERE orbit_id = $1 AND user_id = $2 AND client_id = $3
	`

	revokeConsentSQL = `
		UPDATE consents
		SET revoked = true, updated_at = $2
		WHERE id = $1 AND revoked = false
		RETURNING updated_at
	`
)

func (r *ConsentRepository) Create(ctx context.Context, c *models.Consent) (*models.Consent, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, insertConsentSQL,
		c.OrbitID,
		c.UserID,
		c.ClientID,
		c.Scopes,
		c.GrantedAt,
		c.ExpiresAt,
		c.Revoked,
		now,
		now,
	)

	if err := row.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt); err != nil {
		r.logger.Error().Err(err).Msg("consent create failed")
		return nil, err
	}
	return c, nil
}

func (r *ConsentRepository) Get(ctx context.Context, orbitID, userID, clientID int64) (*models.Consent, error) {
	ctx, span := r.tracer.Start(ctx, "Get")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectConsentSQL, orbitID, userID, clientID)
	return scanConsent(row)
}

func (r *ConsentRepository) Revoke(ctx context.Context, consentID int64) error {
	ctx, span := r.tracer.Start(ctx, "Revoke")
	defer span.End()

	row := r.exec.QueryRow(ctx, revokeConsentSQL, consentID, time.Now().UTC())
	var updatedAt time.Time
	if err := row.Scan(&updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func scanConsent(scanner interface{ Scan(dest ...any) error }) (*models.Consent, error) {
	c := &models.Consent{}
	err := scanner.Scan(
		&c.ID,
		&c.OrbitID,
		&c.UserID,
		&c.ClientID,
		&c.Scopes,
		&c.GrantedAt,
		&c.ExpiresAt,
		&c.Revoked,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return c, err
}
