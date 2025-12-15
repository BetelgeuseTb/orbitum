package postgres

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/db"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type OrbitRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewOrbitRepository(exec db.Executor, logger zerolog.Logger) *OrbitRepository {
	return &OrbitRepository{exec: exec, logger: logger, tracer: otel.Tracer("repository.orbit")}
}

const (
	insertOrbitSQL = `
		insert into orbits (name, description, metadata, created_at, updated_at)
		values ($1, $2, $3, $4, $5)
		returning id, created_at, updated_at
	`

	selectOrbitByIDSQL = `
		select id, name, description, metadata, created_at, updated_at, deleted_at
		from orbits
		where id = $1 and deleted_at is null
	`

	updateOrbitSQL = `
		update orbits
		set name = $2,
		    description = $3,
		    metadata = $4,
		    updated_at = $5
		where id = $1 and deleted_at is null
		returning updated_at
	`

	softDeleteOrbitSQL = `
		update orbits
		set deleted_at = $2
		where id = $1 and deleted_at is null
	`

	listOrbitsSQL = `
		select id, name, description, metadata, created_at, updated_at, deleted_at
		from orbits
		where deleted_at is null
		order by id
		limit $1 offset $2
	`
)

func (r *OrbitRepository) Create(ctx context.Context, orbit *models.Orbit) (*models.Orbit, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, insertOrbitSQL, orbit.Name, orbit.Description, now, now)

	if err := row.Scan(&orbit.ID, &orbit.CreatedAt, &orbit.UpdatedAt); err != nil {
		r.logger.Error().Err(err).Msg("orbit create failed")
		return nil, err
	}

	return orbit, nil
}

func (r *OrbitRepository) GetByID(ctx context.Context, id int64) (*models.Orbit, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectOrbitByIDSQL, id)

	orbit := &models.Orbit{}
	err := row.Scan(
		&orbit.ID,
		&orbit.Name,
		&orbit.Description,
		&orbit.CreatedAt,
		&orbit.UpdatedAt,
		&orbit.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error().Err(err).Int64("orbit_id", id).Msg("orbit get failed")
		return nil, err
	}

	return orbit, nil
}

func (r *OrbitRepository) Update(ctx context.Context, orbit *models.Orbit) (*models.Orbit, error) {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, updateOrbitSQL, orbit.ID, orbit.Name, orbit.Description, now)

	if err := row.Scan(&orbit.UpdatedAt); err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		r.logger.Error().Err(err).Int64("orbit_id", orbit.ID).Msg("orbit update failed")
		return nil, err
	}

	return orbit, nil
}

func (r *OrbitRepository) Delete(ctx context.Context, id int64) error {
	ctx, span := r.tracer.Start(ctx, "Delete")
	defer span.End()

	_, err := r.exec.Exec(ctx, softDeleteOrbitSQL, id, time.Now().UTC())
	if err != nil {
		r.logger.Error().Err(err).Int64("orbit_id", id).Msg("orbit delete failed")
	}
	return err
}

func (r *OrbitRepository) List(ctx context.Context, limit, offset int) ([]*models.Orbit, error) {
	ctx, span := r.tracer.Start(ctx, "List")
	defer span.End()

	rows, err := r.exec.Query(ctx, listOrbitsSQL, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Msg("orbit list failed")
		return nil, err
	}
	defer rows.Close()

	var result []*models.Orbit
	for rows.Next() {
		o := &models.Orbit{}
		if err := rows.Scan(
			&o.ID,
			&o.Name,
			&o.Description,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.DeletedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}
