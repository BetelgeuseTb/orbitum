package repository

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
)

type RevokedTokenRepository interface {
	Add(ctx context.Context, t *model.RevokedToken) error
	IsRevoked(ctx context.Context, jti model.UUID) (bool, error)
}
