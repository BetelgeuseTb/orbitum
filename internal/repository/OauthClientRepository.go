package repository

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
)

type OAuthClientRepository interface {
	GetByID(ctx context.Context, id model.UUID) (*model.OAuthClient, error)
	Create(ctx context.Context, c *model.OAuthClient) error
	Update(ctx context.Context, c *model.OAuthClient) error
	Delete(ctx context.Context, id model.UUID) error
}
