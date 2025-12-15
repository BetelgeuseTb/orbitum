package postgres

import (
	"context"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/db"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type SessionRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewSessionRepository(exec db.Executor, logger zerolog.Logger) *SessionRepository {
	return &SessionRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.session"),
	}
}

const (
	insertSessionSQL = `
		INSERT INTO sessions
			(orbit_id, user_id, client_id, started_at, last_active_at, expires_at, revoked, device_info, ip, metadata, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at, updated_at
	`

	selectSessionByIDSQL = `
		SELECT id, orbit_id, user_id, client_id, started_at, last_active_at, expires_at, revoked, device_info, ip, metadata, created_at, updated_at
		FROM sessions
		WHERE id = $1
		LIMIT 1
	`

	updateSessionSQL = `
		UPDATE sessions
		SET last_active_at = $2, expires_at = $3, revoked = $4, metadata = $5, updated_at = $6
		WHERE id = $1
		RETURNING id, updated_at
	`

	revokeSessionSQL = `
		UPDATE sessions
		SET revoked = TRUE, updated_at = $2
		WHERE id = $1
		RETURNING id
	`

	listSessionsByUserSQL = `
		SELECT id, orbit_id, user_id, client_id, started_at, last_active_at, expires_at, revoked, device_info, ip, metadata, created_at, updated_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
	`
)

func (r *SessionRepository) Create(ctx context.Context, s *models.Session) (*models.Session, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now

	row := r.exec.QueryRow(ctx, insertSessionSQL,
		s.OrbitID,
		s.UserID,
		s.ClientID,
		s.StartedAt,
		s.LastActiveAt,
		s.ExpiresAt,
		s.Revoked,
		s.DeviceInfo,
		s.IP,
		s.Metadata,
		s.CreatedAt,
		s.UpdatedAt,
	)

	if err := row.Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt); err != nil {
		r.logger.Error().Err(err).Int64("user_id", s.UserID).Msg("session create failed")
		return nil, err
	}
	return s, nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id int64) (*models.Session, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectSessionByIDSQL, id)
	s := &models.Session{}
	if err := row.Scan(
		&s.ID,
		&s.OrbitID,
		&s.UserID,
		&s.ClientID,
		&s.StartedAt,
		&s.LastActiveAt,
		&s.ExpiresAt,
		&s.Revoked,
		&s.DeviceInfo,
		&s.IP,
		&s.Metadata,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("session_id", id).Msg("get session by id failed")
		return nil, err
	}
	return s, nil
}

func (r *SessionRepository) Update(ctx context.Context, s *models.Session) (*models.Session, error) {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, updateSessionSQL,
		s.ID,
		s.LastActiveAt,
		s.ExpiresAt,
		s.Revoked,
		s.Metadata,
		now,
	)

	var returnedID int64
	var returnedUpdatedAt time.Time
	if err := row.Scan(&returnedID, &returnedUpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("session_id", s.ID).Msg("update session failed")
		return nil, err
	}
	s.UpdatedAt = returnedUpdatedAt
	s.ID = returnedID
	return s, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id int64) error {
	ctx, span := r.tracer.Start(ctx, "Revoke")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, revokeSessionSQL, id, now)
	var returnedID int64
	if err := row.Scan(&returnedID); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		r.logger.Error().Err(err).Int64("session_id", id).Msg("revoke session failed")
		return err
	}
	return nil
}

func (r *SessionRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*models.Session, error) {
	ctx, span := r.tracer.Start(ctx, "ListByUser")
	defer span.End()

	rows, err := r.exec.Query(ctx, listSessionsByUserSQL, userID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Int64("user_id", userID).Msg("list sessions query failed")
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		s := &models.Session{}
		if err := rows.Scan(
			&s.ID,
			&s.OrbitID,
			&s.UserID,
			&s.ClientID,
			&s.StartedAt,
			&s.LastActiveAt,
			&s.ExpiresAt,
			&s.Revoked,
			&s.DeviceInfo,
			&s.IP,
			&s.Metadata,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}
