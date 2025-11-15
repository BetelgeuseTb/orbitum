package postgres

import (
	"context"
	"errors"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/model"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository"
	sqlq "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/sql"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepoPG struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repo.UserRepository {
	return &userRepoPG{db: db}
}

var ErrNoRows = errors.New("no rows")

func (r *userRepoPG) Create(ctx context.Context, u *model.User) error {
	row := r.db.QueryRow(ctx, sqlq.UserInsert, u.Email, u.PasswordHash, u.IsActive)
	var out model.User
	if err := row.Scan(&out.ID, &out.Email, &out.PasswordHash, &out.IsActive, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return err
	}
	*u = out
	return nil
}

func (r *userRepoPG) GetByID(ctx context.Context, id model.UUID) (*model.User, error) {
	row := r.db.QueryRow(ctx, sqlq.UserGetByID, id)
	var u model.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepoPG) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRow(ctx, sqlq.UserGetByEmail, email)
	var u model.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepoPG) Update(ctx context.Context, u *model.User) error {
	row := r.db.QueryRow(ctx, sqlq.UserUpdate, u.ID, u.Email, u.PasswordHash, u.IsActive)
	var out model.User
	if err := row.Scan(&out.ID, &out.Email, &out.PasswordHash, &out.IsActive, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return err
	}
	*u = out
	return nil
}

func (r *userRepoPG) Delete(ctx context.Context, id model.UUID) error {
	ct, err := r.db.Exec(ctx, sqlq.UserDelete, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNoRows
	}
	return nil
}
