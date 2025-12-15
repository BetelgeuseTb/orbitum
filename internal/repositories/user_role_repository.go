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

type UserRoleRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewUserRoleRepository(exec db.Executor, logger zerolog.Logger) *UserRoleRepository {
	return &UserRoleRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.user_role"),
	}
}

const (
	insertUserRoleSQL = `
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, $3)
	`

	deleteUserRoleSQL = `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = $2
	`

	listRolesByUserSQL = `
		SELECT role_id, created_at
		FROM user_roles
		WHERE user_id = $1
		ORDER BY role_id
		LIMIT $2 OFFSET $3
	`
)

func (r *UserRoleRepository) Assign(ctx context.Context, userID, roleID int64) error {
	ctx, span := r.tracer.Start(ctx, "Assign")
	defer span.End()

	_, err := r.exec.Exec(ctx, insertUserRoleSQL, userID, roleID, time.Now().UTC())
	return err
}

func (r *UserRoleRepository) Revoke(ctx context.Context, userID, roleID int64) error {
	ctx, span := r.tracer.Start(ctx, "Revoke")
	defer span.End()

	_, err := r.exec.Exec(ctx, deleteUserRoleSQL, userID, roleID)
	return err
}

func (r *UserRoleRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*models.UserRole, error) {
	ctx, span := r.tracer.Start(ctx, "ListByUser")
	defer span.End()

	rows, err := r.exec.Query(ctx, listRolesByUserSQL, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.UserRole
	for rows.Next() {
		ur := &models.UserRole{UserID: userID}
		if err := rows.Scan(&ur.RoleID, &ur.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, ur)
	}
	return result, rows.Err()
}
