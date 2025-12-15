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

type TOTPRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewTOTPRepository(exec db.Executor, logger zerolog.Logger) *TOTPRepository {
	return &TOTPRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.totp"),
	}
}

const (
	insertTOTPSQL = `
		INSERT INTO totps
			(created_at, updated_at, user_id, orbit_id, secret_cipher, algorithm, digits, period, issuer, label, last_used_step, is_confirmed, name)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, created_at, updated_at
	`

	selectTOTPByIDSQL = `
		SELECT id, user_id, orbit_id, secret_cipher, algorithm, digits, period, issuer, label, last_used_step, is_confirmed, name, created_at, updated_at
		FROM totps
		WHERE id = $1
		LIMIT 1
	`

	updateTOTPSQL = `
		UPDATE totps
		SET secret_cipher = $2, algorithm = $3, digits = $4, period = $5, issuer = $6, label = $7, last_used_step = $8, is_confirmed = $9, name = $10, updated_at = $11
		WHERE id = $1
		RETURNING id, updated_at
	`

	deleteTOTPSQL = `
		DELETE FROM totps
		WHERE id = $1
		RETURNING id
	`

	listTOTPsByUserSQL = `
		SELECT id, user_id, orbit_id, secret_cipher, algorithm, digits, period, issuer, label, last_used_step, is_confirmed, name, created_at, updated_at
		FROM totps
		WHERE user_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
	`
)

func (r *TOTPRepository) Create(ctx context.Context, t *models.TOTP) (*models.TOTP, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now

	row := r.exec.QueryRow(ctx, insertTOTPSQL,
		t.CreatedAt,
		t.UpdatedAt,
		t.UserID,
		t.OrbitID,
		t.SecretCipher,
		t.Algorithm,
		t.Digits,
		t.Period,
		t.Issuer,
		t.Label,
		t.LastUsedStep,
		t.IsConfirmed,
		t.Name,
	)

	if err := row.Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		r.logger.Error().Err(err).Int64("user_id", t.UserID).Msg("totp create failed")
		return nil, err
	}

	r.logger.Debug().Int64("totp_id", t.ID).Int64("user_id", t.UserID).Msg("totp created")
	return t, nil
}

func (r *TOTPRepository) GetByID(ctx context.Context, id int64) (*models.TOTP, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectTOTPByIDSQL, id)
	t, err := scanTOTPRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("totp_id", id).Msg("totp get failed")
		return nil, err
	}
	return t, nil
}

func (r *TOTPRepository) Update(ctx context.Context, t *models.TOTP) (*models.TOTP, error) {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, updateTOTPSQL,
		t.ID,
		t.SecretCipher,
		t.Algorithm,
		t.Digits,
		t.Period,
		t.Issuer,
		t.Label,
		t.LastUsedStep,
		t.IsConfirmed,
		t.Name,
		now,
	)

	var returnedID int64
	var returnedUpdatedAt time.Time
	if err := row.Scan(&returnedID, &returnedUpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("totp_id", t.ID).Msg("totp update failed")
		return nil, err
	}

	t.ID = returnedID
	t.UpdatedAt = returnedUpdatedAt
	r.logger.Debug().Int64("totp_id", t.ID).Int64("user_id", t.UserID).Msg("totp updated")
	return t, nil
}

func (r *TOTPRepository) Delete(ctx context.Context, id int64) error {
	ctx, span := r.tracer.Start(ctx, "Delete")
	defer span.End()

	row := r.exec.QueryRow(ctx, deleteTOTPSQL, id)
	var returnedID int64
	if err := row.Scan(&returnedID); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		r.logger.Error().Err(err).Int64("totp_id", id).Msg("totp delete failed")
		return err
	}
	r.logger.Info().Int64("totp_id", returnedID).Msg("totp deleted")
	return nil
}

func (r *TOTPRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*models.TOTP, error) {
	ctx, span := r.tracer.Start(ctx, "ListByUser")
	defer span.End()

	rows, err := r.exec.Query(ctx, listTOTPsByUserSQL, userID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Int64("user_id", userID).Msg("totp list failed")
		return nil, err
	}
	defer rows.Close()

	var result []*models.TOTP
	for rows.Next() {
		t, err := scanTOTPRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func scanTOTPRow(scanner interface{ Scan(dest ...any) error }) (*models.TOTP, error) {
	t := &models.TOTP{}
	err := scanner.Scan(
		&t.ID,
		&t.UserID,
		&t.OrbitID,
		&t.SecretCipher,
		&t.Algorithm,
		&t.Digits,
		&t.Period,
		&t.Issuer,
		&t.Label,
		&t.LastUsedStep,
		&t.IsConfirmed,
		&t.Name,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}
