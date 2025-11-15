package repository

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id model.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id model.UUID) error
}
