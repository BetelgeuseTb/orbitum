package repository

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
)

type SessionRepository interface {
	Create(ctx context.Context, s *model.Session) error
	GetByID(ctx context.Context, id model.UUID) (*model.Session, error)
	GetByUser(ctx context.Context, userID model.UUID) ([]model.Session, error)
	Revoke(ctx context.Context, id model.UUID) error
	DeleteExpired(ctx context.Context) error
}
