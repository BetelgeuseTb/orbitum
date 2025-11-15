package repository

import (
	"context"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
)

type AuthorizationCodeRepository interface {
	Create(ctx context.Context, ac *model.AuthorizationCode) error
	Get(ctx context.Context, code string) (*model.AuthorizationCode, error)
	MarkUsed(ctx context.Context, code string) error
	DeleteExpired(ctx context.Context) error
}
