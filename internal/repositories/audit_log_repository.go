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

type AuditLogRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewAuditLogRepository(exec db.Executor, logger zerolog.Logger) *AuditLogRepository {
	return &AuditLogRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.audit_log"),
	}
}

const (
	insertAuditLogSQL = `
		INSERT INTO audit_logs (
			orbit_id, actor_id, action, target, metadata, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id
	`

	listAuditLogsByOrbitSQL = `
		SELECT id, orbit_id, actor_id, action, target, metadata, created_at
		FROM audit_logs
		WHERE orbit_id = $1
		ORDER BY id DESC
		LIMIT $2 OFFSET $3
	`
)

func (r *AuditLogRepository) Create(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	row := r.exec.QueryRow(ctx, insertAuditLogSQL,
		log.OrbitID,
		log.ActorID,
		log.Action,
		log.Target,
		log.Metadata,
		time.Now().UTC(),
	)

	if err := row.Scan(&log.ID); err != nil {
		r.logger.Error().Err(err).Msg("audit log create failed")
		return nil, err
	}
	return log, nil
}

func (r *AuditLogRepository) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.AuditLog, error) {
	ctx, span := r.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	rows, err := r.exec.Query(ctx, listAuditLogsByOrbitSQL, orbitID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.AuditLog
	for rows.Next() {
		a := &models.AuditLog{}
		if err := rows.Scan(
			&a.ID,
			&a.OrbitID,
			&a.ActorID,
			&a.Action,
			&a.Target,
			&a.Metadata,
			&a.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}
