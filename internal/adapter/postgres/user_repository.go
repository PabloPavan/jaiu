package postgres

import (
	"context"
	"errors"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres/sqlc"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{queries: sqlc.New(pool)}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	params := sqlc.CreateUserParams{
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
		Active:       user.Active,
	}

	created, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return domain.User{}, err
	}

	return mapUser(created), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ports.ErrNotFound
		}
		return domain.User{}, err
	}

	return mapUser(user), nil
}

func mapUser(user sqlc.User) domain.User {
	return domain.User{
		ID:           uuidToString(user.ID),
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         domain.UserRole(user.Role),
		Active:       user.Active,
		CreatedAt:    timeFrom(user.CreatedAt),
		UpdatedAt:    timeFrom(user.UpdatedAt),
	}
}
