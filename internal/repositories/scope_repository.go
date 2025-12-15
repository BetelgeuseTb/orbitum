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

type ScopeRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewScopeRepository(exec db.Executor, logger zerolog.Logger) *ScopeRepository {
	return &ScopeRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.scope"),
	}
}

const (
	insertScopeSQL = `
		INSERT INTO scopes (orbit_id, name, description, is_default, is_active, metadata, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, created_at, updated_at
	`

	selectScopeByIDSQL = `
		SELECT id, orbit_id, name, description, is_default, is_active, metadata, created_at, updated_at, deleted_at
		FROM scopes
		WHERE id = $1 AND deleted_at IS NULL
	`

	selectScopeByNameSQL = `
		SELECT id, orbit_id, name, description, is_default, is_active, metadata, created_at, updated_at, deleted_at
		FROM scopes
		WHERE orbit_id = $1 AND name = $2 AND deleted_at IS NULL
	`

	updateScopeSQL = `
		UPDATE scopes
		SET description = $2, is_default = $3, is_active = $4, metadata = $5, updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at
	`

	softDeleteScopeSQL = `
		UPDATE scopes
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`

	listScopesByOrbitSQL = `
		SELECT id, orbit_id, name, description, is_default, is_active, metadata, created_at, updated_at, deleted_at
		FROM scopes
		WHERE orbit_id = $1 AND deleted_at IS NULL
		ORDER BY id
		LIMIT $2 OFFSET $3
	`
)

func (r *ScopeRepository) Create(ctx context.Context, s *models.Scope) (*models.Scope, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, insertScopeSQL,
		s.OrbitID,
		s.Name,
		s.Description,
		s.IsDefault,
		s.IsActive,
		s.Metadata,
		now,
		now,
	)
	if err := row.Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt); err != nil {
		r.logger.Error().Err(err).Str("name", s.Name).Int64("orbit_id", s.OrbitID).Msg("scope create failed")
		return nil, err
	}
	return s, nil
}

func (r *ScopeRepository) GetByID(ctx context.Context, id int64) (*models.Scope, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectScopeByIDSQL, id)
	return scanScopeRow(row)
}

func (r *ScopeRepository) GetByName(ctx context.Context, orbitID int64, name string) (*models.Scope, error) {
	ctx, span := r.tracer.Start(ctx, "GetByName")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectScopeByNameSQL, orbitID, name)
	return scanScopeRow(row)
}

func (r *ScopeRepository) Update(ctx context.Context, s *models.Scope) (*models.Scope, error) {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, updateScopeSQL,
		s.ID,
		s.Description,
		s.IsDefault,
		s.IsActive,
		s.Metadata,
		now,
	)
	if err := row.Scan(&s.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("scope_id", s.ID).Msg("scope update failed")
		return nil, err
	}
	return s, nil
}

func (r *ScopeRepository) SoftDelete(ctx context.Context, id int64) error {
	ctx, span := r.tracer.Start(ctx, "SoftDelete")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, softDeleteScopeSQL, id, now)
	var returnedID int64
	if err := row.Scan(&returnedID); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		r.logger.Error().Err(err).Int64("scope_id", id).Msg("scope soft delete failed")
		return err
	}
	return nil
}

func (r *ScopeRepository) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.Scope, error) {
	ctx, span := r.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	rows, err := r.exec.Query(ctx, listScopesByOrbitSQL, orbitID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Int64("orbit_id", orbitID).Msg("list scopes query failed")
		return nil, err
	}
	defer rows.Close()

	var result []*models.Scope
	for rows.Next() {
		s, err := scanScopeRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func scanScopeRow(scanner interface{ Scan(dest ...any) error }) (*models.Scope, error) {
	s := &models.Scope{}
	err := scanner.Scan(
		&s.ID,
		&s.OrbitID,
		&s.Name,
		&s.Description,
		&s.IsDefault,
		&s.IsActive,
		&s.Metadata,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}
