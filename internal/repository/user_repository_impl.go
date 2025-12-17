package repository

import (
	"context"
	database "user-api/db/sqlc"
)

type UserRepositoryImpl struct {
	queries *database.Queries
}

func NewUserRepository(queries *database.Queries) UserRepository {
	return &UserRepositoryImpl{
		queries: queries,
	}
}

func (r *UserRepositoryImpl) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return r.queries.CreateUser(ctx, arg)
}

func (r *UserRepositoryImpl) GetUser(ctx context.Context, id int32) (database.User, error) {
	return r.queries.GetUser(ctx, id)
}

func (r *UserRepositoryImpl) ListUsers(ctx context.Context) ([]database.User, error) {
	return r.queries.ListUsers(ctx)
}

func (r *UserRepositoryImpl) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return r.queries.UpdateUser(ctx, arg)
}

func (r *UserRepositoryImpl) DeleteUser(ctx context.Context, id int32) error {
	_, err := r.queries.DeleteUser(ctx, id)
	return err
}
