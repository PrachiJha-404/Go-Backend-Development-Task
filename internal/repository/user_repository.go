package repository

import (
	"context"
	database "user-api/db/sqlc"
)

type UserRepository interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	GetUser(ctx context.Context, id int32) (database.User, error)
	ListUsers(ctx context.Context) ([]database.User, error)
	UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error)
	DeleteUser(ctx context.Context, id int32) error
}

