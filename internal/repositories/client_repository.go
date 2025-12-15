package postgres

import (
	"context"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/models"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repositories/db"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ClientRepository struct {
	exec   db.Executor
	logger zerolog.Logger
	tracer trace.Tracer
}

func NewClientRepository(exec db.Executor, logger zerolog.Logger) *ClientRepository {
	return &ClientRepository{
		exec:   exec,
		logger: logger,
		tracer: otel.Tracer("repository.client"),
	}
}

const (
	insertClientSQL = `
		INSERT INTO clients (
			orbit_id, client_id, client_secret_hash, name, description,
			redirect_uris, post_logout_redirect_uris, grant_types, response_types,
			token_endpoint_auth_method, contacts, logo_uri, app_type, is_public,
			is_active, allowed_cors_origins, allowed_scopes, metadata, created_at, updated_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
		)
		RETURNING id, created_at, updated_at, deleted_at
	`

	selectClientByIDSQL = `
		SELECT id, orbit_id, client_id, client_secret_hash, name, description,
		       redirect_uris, post_logout_redirect_uris, grant_types, response_types,
		       token_endpoint_auth_method, contacts, logo_uri, app_type, is_public,
		       is_active, allowed_cors_origins, allowed_scopes, metadata, created_at, updated_at, deleted_at
		FROM clients
		WHERE id = $1 AND deleted_at IS NULL
	`

	selectClientByClientIDSQL = `
		SELECT id, orbit_id, client_id, client_secret_hash, name, description,
		       redirect_uris, post_logout_redirect_uris, grant_types, response_types,
		       token_endpoint_auth_method, contacts, logo_uri, app_type, is_public,
		       is_active, allowed_cors_origins, allowed_scopes, metadata, created_at, updated_at, deleted_at
		FROM clients
		WHERE orbit_id = $1 AND client_id = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	updateClientSQL = `
		UPDATE clients
		SET client_secret_hash = $3, name = $4, description = $5,
		    redirect_uris = $6, post_logout_redirect_uris = $7, grant_types = $8, response_types = $9,
		    token_endpoint_auth_method = $10, contacts = $11, logo_uri = $12, app_type = $13,
		    is_public = $14, is_active = $15, allowed_cors_origins = $16, allowed_scopes = $17,
		    metadata = $18, updated_at = $19
		WHERE id = $1 AND orbit_id = $2 AND deleted_at IS NULL
		RETURNING updated_at
	`

	softDeleteClientSQL = `
		UPDATE clients
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`

	listClientsByOrbitSQL = `
		SELECT id, orbit_id, client_id, client_secret_hash, name, description,
		       redirect_uris, post_logout_redirect_uris, grant_types, response_types,
		       token_endpoint_auth_method, contacts, logo_uri, app_type, is_public,
		       is_active, allowed_cors_origins, allowed_scopes, metadata, created_at, updated_at, deleted_at
		FROM clients
		WHERE orbit_id = $1 AND deleted_at IS NULL
		ORDER BY id
		LIMIT $2 OFFSET $3
	`
)

func (r *ClientRepository) Create(ctx context.Context, c *models.Client) (*models.Client, error) {
	ctx, span := r.tracer.Start(ctx, "Create")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, insertClientSQL,
		c.OrbitID,
		c.ClientID,
		c.ClientSecretHash,
		c.Name,
		c.Description,
		c.RedirectURIs,
		c.PostLogoutRedirectURIs,
		c.GrantTypes,
		c.ResponseTypes,
		c.TokenEndpointAuthMethod,
		c.Contacts,
		c.LogoURI,
		c.AppType,
		c.IsPublic,
		c.IsActive,
		c.AllowedCORSOrigins,
		c.AllowedScopes,
		c.Metadata,
		now,
		now,
	)

	var deletedAt *time.Time
	if err := row.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt, &deletedAt); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			r.logger.Warn().Err(err).Str("client_id", c.ClientID).Msg("unique constraint violation on create client")
			return nil, err
		}
		r.logger.Error().Err(err).Str("client_id", c.ClientID).Msg("create client failed")
		return nil, err
	}
	c.DeletedAt = deletedAt
	r.logger.Info().Int64("client_id", c.ID).Str("client_client_id", c.ClientID).Msg("client created")
	return c, nil
}

func (r *ClientRepository) GetByID(ctx context.Context, id int64) (*models.Client, error) {
	ctx, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectClientByIDSQL, id)
	return scanClientRow(row)
}

func (r *ClientRepository) GetByClientID(ctx context.Context, orbitID int64, clientID string) (*models.Client, error) {
	ctx, span := r.tracer.Start(ctx, "GetByClientID")
	defer span.End()

	row := r.exec.QueryRow(ctx, selectClientByClientIDSQL, orbitID, clientID)
	return scanClientRow(row)
}

func (r *ClientRepository) Update(ctx context.Context, c *models.Client) (*models.Client, error) {
	ctx, span := r.tracer.Start(ctx, "Update")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, updateClientSQL,
		c.ID,
		c.OrbitID,
		c.ClientSecretHash,
		c.Name,
		c.Description,
		c.RedirectURIs,
		c.PostLogoutRedirectURIs,
		c.GrantTypes,
		c.ResponseTypes,
		c.TokenEndpointAuthMethod,
		c.Contacts,
		c.LogoURI,
		c.AppType,
		c.IsPublic,
		c.IsActive,
		c.AllowedCORSOrigins,
		c.AllowedScopes,
		c.Metadata,
		now,
	)

	var updatedAt time.Time
	if err := row.Scan(&updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			r.logger.Warn().Err(err).Int64("client_db_id", c.ID).Msg("unique constraint violation on update client")
			return nil, err
		}
		r.logger.Error().Err(err).Int64("client_db_id", c.ID).Msg("update client failed")
		return nil, err
	}
	c.UpdatedAt = updatedAt
	r.logger.Debug().Int64("client_db_id", c.ID).Str("client_id", c.ClientID).Msg("client updated")
	return c, nil
}

func (r *ClientRepository) Delete(ctx context.Context, id int64) error {
	ctx, span := r.tracer.Start(ctx, "Delete")
	defer span.End()

	now := time.Now().UTC()
	row := r.exec.QueryRow(ctx, softDeleteClientSQL, id, now)
	var deletedID int64
	if err := row.Scan(&deletedID); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		r.logger.Error().Err(err).Int64("client_id", id).Msg("soft delete client failed")
		return err
	}
	r.logger.Info().Int64("client_id", id).Msg("client soft-deleted")
	return nil
}

func (r *ClientRepository) ListByOrbit(ctx context.Context, orbitID int64, limit, offset int) ([]*models.Client, error) {
	ctx, span := r.tracer.Start(ctx, "ListByOrbit")
	defer span.End()

	rows, err := r.exec.Query(ctx, listClientsByOrbitSQL, orbitID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Int64("orbit_id", orbitID).Msg("list clients query failed")
		return nil, err
	}
	defer rows.Close()

	var clients []*models.Client
	for rows.Next() {
		c, err := scanClientRow(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clients, nil
}

func scanClientRow(scanner interface{ Scan(dest ...any) error }) (*models.Client, error) {
	c := &models.Client{}
	err := scanner.Scan(
		&c.ID,
		&c.OrbitID,
		&c.ClientID,
		&c.ClientSecretHash,
		&c.Name,
		&c.Description,
		&c.RedirectURIs,
		&c.PostLogoutRedirectURIs,
		&c.GrantTypes,
		&c.ResponseTypes,
		&c.TokenEndpointAuthMethod,
		&c.Contacts,
		&c.LogoURI,
		&c.AppType,
		&c.IsPublic,
		&c.IsActive,
		&c.AllowedCORSOrigins,
		&c.AllowedScopes,
		&c.Metadata,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return c, err
}
