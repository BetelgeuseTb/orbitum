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

type RoleRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewRoleRepository(exec db.Executor, logger zerolog.Logger) *RoleRepository {
	return &RoleRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.role"),
	}
}

const (
	insertRoleSQL = `
		INSERT INTO roles (orbit_id, name, metadata, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	selectRoleByIDSQL = `
		SELECT id, orbit_id, name, metadata, created_at
		FROM roles
		WHERE id = $1
	`

	listRolesByOrbitSQL = `
		SELECT id, orbit_id, name, metadata, created_at
		FROM roles
		WHERE orbit_id = $1
		ORDER BY id
		LIMIT $2 OFFSET $3
	`

	deleteRoleSQL = `
		DELETE FROM roles
		WHERE id = $1
	`
)

func (r *RoleRepository) Create(ctx context.Context, role *models.Role) (*models.Role, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, insertRoleSQL,
		role.OrbitID,
		role.Name,
		role.Metadata,
		now,
	)

	if err := row.Scan(&role.ID, &role.CreatedAt); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id int64) (*models.Role, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectRoleByIDSQL, id)
	return scanRole(row)
}

func (r *RoleRepository) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.Role, error) {
	ctx, span := r.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	rows, err := r.exec.Query(ctx, listRolesByOrbitSQL, orbitID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, role)
	}
	return result, rows.Err()
}

func (r *RoleRepository) Delete(ctx context.Context, id int64) error {
	ctx, span := r.tracer.Start(ctx, "Delete")
	defer span.End()

	_, err := r.exec.Exec(ctx, deleteRoleSQL, id)
	return err
}

func scanRole(scanner interface{ Scan(dest ...any) error }) (*models.Role, error) {
	r := &models.Role{}
	err := scanner.Scan(
		&r.ID,
		&r.OrbitID,
		&r.Name,
		&r.Metadata,
		&r.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return r, err
}
