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

type DeviceCodeRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewDeviceCodeRepository(exec db.Executor, logger zerolog.Logger) *DeviceCodeRepository {
	return &DeviceCodeRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.device_code"),
	}
}

const (
	insertDeviceCodeSQL = `
		INSERT INTO device_codes (
			orbit_id, client_id, device_code_hash, user_code, scopes,
			expires_at, poll_interval_sec, status, user_id, metadata,
			created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at, updated_at
	`

	selectDeviceCodeByUserCodeSQL = `
		SELECT id, orbit_id, client_id, device_code_hash, user_code, scopes,
		       expires_at, poll_interval_sec, status, user_id, metadata,
		       created_at, updated_at
		FROM device_codes
		WHERE user_code = $1 AND orbit_id = $2
		LIMIT 1
	`

	updateDeviceCodeStatusSQL = `
		UPDATE device_codes
		SET status = $2, user_id = $3, updated_at = $4
		WHERE id = $1
		RETURNING updated_at
	`
)

func (r *DeviceCodeRepository) Create(ctx context.Context, dc *models.DeviceCode) (*models.DeviceCode, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(
		ctx,
		insertDeviceCodeSQL,
		dc.OrbitID,
		dc.ClientID,
		dc.DeviceCodeHash,
		dc.UserCode,
		dc.Scopes,
		dc.ExpiresAt,
		dc.PollIntervalSec,
		dc.Status,
		dc.UserID,
		dc.Metadata,
		now,
		now,
	)

	if err := row.Scan(&dc.ID, &dc.CreatedAt, &dc.UpdatedAt); err != nil {
		r.logger.Error().Err(err).Str("user_code", dc.UserCode).Msg("device code insert failed")
		return nil, err
	}
	return dc, nil
}

func (r *DeviceCodeRepository) GetByUserCode(ctx context.Context, orbitID int64, userCode string) (*models.DeviceCode, error) {
	ctx, span := r.tracer.Start(ctx, "GetByUserCode")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectDeviceCodeByUserCodeSQL, userCode, orbitID)
	dc := &models.DeviceCode{}
	if err := row.Scan(
		&dc.ID,
		&dc.OrbitID,
		&dc.ClientID,
		&dc.DeviceCodeHash,
		&dc.UserCode,
		&dc.Scopes,
		&dc.ExpiresAt,
		&dc.PollIntervalSec,
		&dc.Status,
		&dc.UserID,
		&dc.Metadata,
		&dc.CreatedAt,
		&dc.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("user_code", userCode).Msg("device code get failed")
		return nil, err
	}
	return dc, nil
}

func (r *DeviceCodeRepository) UpdateStatus(ctx context.Context, id int64, status models.DeviceCodeStatus, userID *int64) error {
	ctx, span := r.tracer.Start(ctx, "UpdateStatus")
	defer span.End()

	row := r.exec.QueryRow(ctx, updateDeviceCodeStatusSQL, id, status, userID, time.Now().UTC())
	var updatedAt time.Time
	if err := row.Scan(&updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		r.logger.Error().Err(err).Int64("device_code_id", id).Msg("update device code status failed")
		return err
	}
	return nil
}
