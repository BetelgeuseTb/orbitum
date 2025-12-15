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

type RolePermissionRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewRolePermissionRepository(exec db.Executor, logger zerolog.Logger) *RolePermissionRepository {
	return &RolePermissionRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.role_permission"),
	}
}

const (
	assignPermissionToRoleSQL = `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`

	revokePermissionFromRoleSQL = `
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2
	`

	listPermissionsByRoleSQL = `
		SELECT p.id, p.orbit_id, p.name, p.metadata, p.created_at
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.id
		LIMIT $2 OFFSET $3
	`
)

func (r *RolePermissionRepository) Assign(ctx context.Context, roleID, permissionID int64) error {
	ctx, span := r.tracer.Start(ctx, "Assign")
	defer span.End()

	_, err := r.exec.Exec(ctx, assignPermissionToRoleSQL, roleID, permissionID, time.Now().UTC())
	if err != nil {
		r.logger.Error().Err(err).Int64("role_id", roleID).Int64("permission_id", permissionID).Msg("assign permission to role failed")
	}
	return err
}

func (r *RolePermissionRepository) Revoke(ctx context.Context, roleID, permissionID int64) error {
	ctx, span := r.tracer.Start(ctx, "Revoke")
	defer span.End()

	_, err := r.exec.Exec(ctx, revokePermissionFromRoleSQL, roleID, permissionID)
	if err != nil {
		r.logger.Error().Err(err).Int64("role_id", roleID).Int64("permission_id", permissionID).Msg("revoke permission from role failed")
	}
	return err
}

func (r *RolePermissionRepository) ListPermissionsByRole(ctx context.Context, roleID int64, limit, offset int) ([]*models.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "ListPermissionsByRole")
	defer span.End()

	rows, err := r.exec.Query(ctx, listPermissionsByRoleSQL, roleID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Int64("role_id", roleID).Msg("list permissions by role query failed")
		return nil, err
	}
	defer rows.Close()

	var result []*models.Permission
	for rows.Next() {
		p := &models.Permission{}
		if err := rows.Scan(&p.ID, &p.OrbitID, &p.Name, &p.Metadata, &p.CreatedAt); err != nil {
			r.logger.Error().Err(err).Int64("role_id", roleID).Msg("scan permission row failed")
			return nil, err
		}
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
