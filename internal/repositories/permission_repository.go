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

type PermissionRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewPermissionRepository(exec db.Executor, logger zerolog.Logger) *PermissionRepository {
	return &PermissionRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.permission"),
	}
}

const (
	insertPermissionSQL = `
		INSERT INTO permissions (orbit_id, name, metadata, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	selectPermissionByIDSQL = `
		SELECT id, orbit_id, name, metadata, created_at
		FROM permissions
		WHERE id = $1
	`

	listPermissionsByOrbitSQL = `
		SELECT id, orbit_id, name, metadata, created_at
		FROM permissions
		WHERE orbit_id = $1
		ORDER BY id
		LIMIT $2 OFFSET $3
	`

	deletePermissionSQL = `
		DELETE FROM permissions
		WHERE id = $1
	`
)

func (r *PermissionRepository) Create(ctx context.Context, permission *models.Permission) (*models.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, insertPermissionSQL,
		permission.OrbitID,
		permission.Name,
		permission.Metadata,
		now,
	)

	if err := row.Scan(&permission.ID, &permission.CreatedAt); err != nil {
		r.logger.Error().Err(err).Str("name", permission.Name).Msg("permission create failed")
		return nil, err
	}
	return permission, nil
}

func (r *PermissionRepository) GetByID(ctx context.Context, id int64) (*models.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectPermissionByIDSQL, id)
	return scanPermission(row)
}

func (r *PermissionRepository) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	rows, err := r.exec.Query(ctx, listPermissionsByOrbitSQL, orbitID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Permission
	for rows.Next() {
		permission, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, permission)
	}
	return result, rows.Err()
}

func (r *PermissionRepository) Delete(ctx context.Context, id int64) error {
	ctx, span := r.tracer.Start(ctx, "Delete")
	defer span.End()

	_, err := r.exec.Exec(ctx, deletePermissionSQL, id)
	return err
}

func scanPermission(scanner interface{ Scan(dest ...any) error }) (*models.Permission, error) {
	p := &models.Permission{}
	err := scanner.Scan(
		&p.ID,
		&p.OrbitID,
		&p.Name,
		&p.Metadata,
		&p.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return p, err
}
