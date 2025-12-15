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

type JWKRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewJWKRepository(exec db.Executor, logger zerolog.Logger) *JWKRepository {
	return &JWKRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.jwk"),
	}
}

const (
	insertJWKSQL = `
		INSERT INTO jwks
			(orbit_id, kid, "use", alg, kty, public_key_jwk, private_key_cipher, is_active, not_before, expires_at, metadata, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, created_at, updated_at
	`

	selectJWKByIDSQL = `
		SELECT id, orbit_id, kid, "use", alg, kty, public_key_jwk, private_key_cipher, is_active, not_before, expires_at, metadata, created_at, updated_at
		FROM jwks
		WHERE id = $1
		LIMIT 1
	`

	selectJWKByOrbitAndKidSQL = `
		SELECT id, orbit_id, kid, "use", alg, kty, public_key_jwk, private_key_cipher, is_active, not_before, expires_at, metadata, created_at, updated_at
		FROM jwks
		WHERE orbit_id = $1 AND kid = $2
		LIMIT 1
	`

	updateJWKSQL = `
		UPDATE jwks
		SET public_key_jwk = $3, private_key_cipher = $4, is_active = $5, not_before = $6, expires_at = $7, metadata = $8, updated_at = $9
		WHERE id = $1 AND orbit_id = $2
		RETURNING id, updated_at
	`

	listJWKsByOrbitSQL = `
		SELECT id, orbit_id, kid, "use", alg, kty, public_key_jwk, private_key_cipher, is_active, not_before, expires_at, metadata, created_at, updated_at
		FROM jwks
		WHERE orbit_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
	`

	deleteJWKSQL = `
		DELETE FROM jwks
		WHERE id = $1 AND orbit_id = $2
		RETURNING id
	`
)

func (r *JWKRepository) Create(ctx context.Context, jwk *models.JWKey) (*models.JWKey, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	if jwk.CreatedAt.IsZero() {
		jwk.CreatedAt = now
	}
	jwk.UpdatedAt = now

	row := r.exec.QueryRow(ctx, insertJWKSQL,
		jwk.OrbitID,
		jwk.Kid,
		jwk.Use,
		jwk.Alg,
		jwk.Kty,
		jwk.PublicKeyJWK,
		jwk.PrivateKeyCipher,
		jwk.IsActive,
		jwk.NotBefore,
		jwk.ExpiresAt,
		jwk.Metadata,
		jwk.CreatedAt,
		jwk.UpdatedAt,
	)

	if err := row.Scan(&jwk.ID, &jwk.CreatedAt, &jwk.UpdatedAt); err != nil {
		r.logger.Error().Err(err).Str("kid", jwk.Kid).Int64("orbit_id", jwk.OrbitID).Msg("jwk create failed")
		return nil, err
	}

	r.logger.Debug().Int64("jwk_id", jwk.ID).Str("kid", jwk.Kid).Msg("jwk created")
	return jwk, nil
}

func (r *JWKRepository) GetByID(ctx context.Context, id int64) (*models.JWKey, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectJWKByIDSQL, id)
	jwk, err := scanJWKRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("jwk_id", id).Msg("jwk get failed")
		return nil, err
	}
	return jwk, nil
}

func (r *JWKRepository) GetByOrbitAndKid(ctx context.Context, orbitID int64, kid string) (*models.JWKey, error) {
	ctx, span := r.tracer.Start(ctx, "GetByOrbitAndKid")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectJWKByOrbitAndKidSQL, orbitID, kid)
	jwk, err := scanJWKRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("orbit_id", orbitID).Str("kid", kid).Msg("jwk get by orbit/kid failed")
		return nil, err
	}
	return jwk, nil
}

func (r *JWKRepository) Update(ctx context.Context, jwk *models.JWKey) (*models.JWKey, error) {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, updateJWKSQL,
		jwk.ID,
		jwk.OrbitID,
		jwk.PublicKeyJWK,
		jwk.PrivateKeyCipher,
		jwk.IsActive,
		jwk.NotBefore,
		jwk.ExpiresAt,
		jwk.Metadata,
		now,
	)

	var returnedID int64
	var returnedUpdatedAt time.Time
	if err := row.Scan(&returnedID, &returnedUpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Int64("jwk_id", jwk.ID).Msg("jwk update failed")
		return nil, err
	}

	jwk.ID = returnedID
	jwk.UpdatedAt = returnedUpdatedAt
	r.logger.Debug().Int64("jwk_id", jwk.ID).Str("kid", jwk.Kid).Msg("jwk updated")
	return jwk, nil
}

func (r *JWKRepository) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.JWKey, error) {
	ctx, span := r.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	rows, err := r.exec.Query(ctx, listJWKsByOrbitSQL, orbitID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Int64("orbit_id", orbitID).Msg("jwk list failed")
		return nil, err
	}
	defer rows.Close()

	var result []*models.JWKey
	for rows.Next() {
		jwk, err := scanJWKRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, jwk)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *JWKRepository) Delete(ctx context.Context, id int64, orbitID int64) error {
	ctx, span := r.tracer.Start(ctx, "Delete")
	defer span.End()

	row := r.exec.QueryRow(ctx, deleteJWKSQL, id, orbitID)
	var returnedID int64
	if err := row.Scan(&returnedID); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		r.logger.Error().Err(err).Int64("jwk_id", id).Int64("orbit_id", orbitID).Msg("jwk delete failed")
		return err
	}
	r.logger.Info().Int64("jwk_id", returnedID).Int64("orbit_id", orbitID).Msg("jwk deleted")
	return nil
}

func scanJWKRow(scanner interface{ Scan(dest ...any) error }) (*models.JWKey, error) {
	j := &models.JWKey{}
	err := scanner.Scan(
		&j.ID,
		&j.OrbitID,
		&j.Kid,
		&j.Use,
		&j.Alg,
		&j.Kty,
		&j.PublicKeyJWK,
		&j.PrivateKeyCipher,
		&j.IsActive,
		&j.NotBefore,
		&j.ExpiresAt,
		&j.Metadata,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return j, nil
}
