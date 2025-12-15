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

type RecoveryCodeRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewRecoveryCodeRepository(exec db.Executor, logger zerolog.Logger) *RecoveryCodeRepository {
	return &RecoveryCodeRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.recovery_code"),
	}
}

const (
	insertRecoveryCodeSQL = `
		INSERT INTO recovery_codes (user_id, code_hash, used_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	selectRecoveryCodeByIDSQL = `
		SELECT id, user_id, code_hash, used_at, created_at
		FROM recovery_codes
		WHERE id = $1
		LIMIT 1
	`

	listRecoveryCodesByUserSQL = `
		SELECT id, user_id, code_hash, used_at, created_at
		FROM recovery_codes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	markRecoveryCodeUsedSQL = `
		UPDATE recovery_codes
		SET used_at = $2
		WHERE id = $1 AND used_at IS NULL
		RETURNING used_at
	`
)

func (r *RecoveryCodeRepository) Create(ctx context.Context, rc *models.RecoveryCode) (*models.RecoveryCode, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, insertRecoveryCodeSQL, rc.UserID, rc.CodeHash, nil, now)
	if err := row.Scan(&rc.ID, &rc.CreatedAt); err != nil {
		r.logger.Error().Err(err).Int64("user_id", rc.UserID).Msg("recovery code insert failed")
		return nil, err
	}
	return rc, nil
}

func (r *RecoveryCodeRepository) GetByID(ctx context.Context, id int64) (*models.RecoveryCode, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectRecoveryCodeByIDSQL, id)
	rc := &models.RecoveryCode{}
	if err := row.Scan(&rc.ID, &rc.UserID, &rc.CodeHash, &rc.UsedAt, &rc.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("recovery_code_id", id).Msg("get recovery code failed")
		return nil, err
	}
	return rc, nil
}

func (r *RecoveryCodeRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*models.RecoveryCode, error) {
	ctx, span := r.tracer.Start(ctx, "ListByUser")
	defer span.End()

	rows, err := r.exec.Query(ctx, listRecoveryCodesByUserSQL, userID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Int64("user_id", userID).Msg("list recovery codes query failed")
		return nil, err
	}
	defer rows.Close()

	var result []*models.RecoveryCode
	for rows.Next() {
		rc := &models.RecoveryCode{}
		if err := rows.Scan(&rc.ID, &rc.UserID, &rc.CodeHash, &rc.UsedAt, &rc.CreatedAt); err != nil {
			r.logger.Error().Err(err).Int64("user_id", userID).Msg("scan recovery code row failed")
			return nil, err
		}
		result = append(result, rc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RecoveryCodeRepository) Use(ctx context.Context, id int64) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "Use")
	defer span.End()

	row := r.exec.QueryRow(ctx, markRecoveryCodeUsedSQL, id, time.Now().UTC())
	var usedAt time.Time
	if err := row.Scan(&usedAt); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		r.logger.Error().Err(err).Int64("recovery_code_id", id).Msg("mark recovery code used failed")
		return false, err
	}
	return true, nil
}
