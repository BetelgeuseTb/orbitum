package repository

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
)

type AccessTokenRepository interface {
	Record(ctx context.Context, t *model.AccessTokenRecord) error
	GetByJTI(ctx context.Context, jti model.UUID) (*model.AccessTokenRecord, error)
	DeleteExpired(ctx context.Context) error
}
