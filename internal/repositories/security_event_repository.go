package postgres

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/db"
)

type SecurityEventRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewSecurityEventRepository(exec db.Executor, logger zerolog.Logger) *SecurityEventRepository {
	return &SecurityEventRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.security_event"),
	}
}

const (
	insertSecurityEventSQL = `
		INSERT INTO security_events (
			orbit_id, user_id, event_type, severity, metadata, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id
	`

	listSecurityEventsByOrbitSQL = `
		SELECT id, orbit_id, user_id, event_type, severity, metadata, created_at
		FROM security_events
		WHERE orbit_id = $1
		ORDER BY id DESC
		LIMIT $2 OFFSET $3
	`
)

func (r *SecurityEventRepository) Create(ctx context.Context, e *models.SecurityEvent) (*models.SecurityEvent, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	row := r.exec.QueryRow(ctx, insertSecurityEventSQL,
		e.OrbitID,
		e.UserID,
		e.EventType,
		e.Severity,
		e.Metadata,
		time.Now().UTC(),
	)

	if err := row.Scan(&e.ID); err != nil {
		r.logger.Error().Err(err).Msg("security event create failed")
		return nil, err
	}
	return e, nil
}

func (r *SecurityEventRepository) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.SecurityEvent, error) {
	ctx, span := r.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	rows, err := r.exec.Query(ctx, listSecurityEventsByOrbitSQL, orbitID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.SecurityEvent
	for rows.Next() {
		e := &models.SecurityEvent{}
		if err := rows.Scan(
			&e.ID,
			&e.OrbitID,
			&e.UserID,
			&e.EventType,
			&e.Severity,
			&e.Metadata,
			&e.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
