package repository

import (
	"context"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
)

type RoleRepository interface {
	GetAll(ctx context.Context) ([]model.Role, error)
	GetByID(ctx context.Context, id int) (*model.Role, error)
	Create(ctx context.Context, r *model.Role) error
	Update(ctx context.Context, r *model.Role) error
	Delete(ctx context.Context, id int) error
}
