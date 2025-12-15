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

type PasswordHistoryRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewPasswordHistoryRepository(exec db.Executor, logger zerolog.Logger) *PasswordHistoryRepository {
	return &PasswordHistoryRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.password_history"),
	}
}

const (
	insertPasswordHistorySQL = `
		INSERT INTO password_history (user_id, password_hash, created_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	selectPasswordHistoryByUserSQL = `
		SELECT id, user_id, password_hash, created_at
		FROM password_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	checkPasswordReuseSQL = `
		SELECT 1
		FROM password_history
		WHERE user_id = $1 AND password_hash = $2
		LIMIT 1
	`
)

func (r *PasswordHistoryRepository) Create(ctx context.Context, ph *models.PasswordHistory) (*models.PasswordHistory, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, insertPasswordHistorySQL, ph.UserID, ph.PasswordHash, now)
	if err := row.Scan(&ph.ID, &ph.CreatedAt); err != nil {
		r.logger.Error().Err(err).Int64("user_id", ph.UserID).Msg("password history insert failed")
		return nil, err
	}
	return ph, nil
}

func (r *PasswordHistoryRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*models.PasswordHistory, error) {
	ctx, span := r.tracer.Start(ctx, "ListByUser")
	defer span.End()

	rows, err := r.exec.Query(ctx, selectPasswordHistoryByUserSQL, userID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Int64("user_id", userID).Msg("password history list query failed")
		return nil, err
	}
	defer rows.Close()

	var result []*models.PasswordHistory
	for rows.Next() {
		ph := &models.PasswordHistory{}
		if err := rows.Scan(&ph.ID, &ph.UserID, &ph.PasswordHash, &ph.CreatedAt); err != nil {
			r.logger.Error().Err(err).Int64("user_id", userID).Msg("scan password history row failed")
			return nil, err
		}
		result = append(result, ph)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PasswordHistoryRepository) ExistsWithHash(ctx context.Context, userID int64, hash string) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "ExistsWithHash")
	defer span.End()

	row := r.exec.QueryRow(ctx, checkPasswordReuseSQL, userID, hash)
	var one int
	if err := row.Scan(&one); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		r.logger.Error().Err(err).Int64("user_id", userID).Msg("password reuse check failed")
		return false, err
	}
	return true, nil
}
