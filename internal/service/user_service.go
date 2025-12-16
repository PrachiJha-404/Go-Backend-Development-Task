package service

import (
	"context"
	"time"
	database "user-api/db/sqlc"
	"user-api/internal/models"
	"user-api/internal/repository"

	"go.uber.org/zap"
)

type UserService struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

func NewUserService(repo repository.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{repo: repo, logger: logger}
}

func (s *UserService) GetUser(ctx context.Context, id int32) (models.UserResponse, error) {
	dbUser, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return models.UserResponse{}, err
	}
	return models.UserResponse{
		ID:   dbUser.ID,
		Name: dbUser.Name,
		DOB:  dbUser.Dob,
		Age:  calculateAge(dbUser.Dob),
	}, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]models.UserResponse, error) {
	userResponse := []models.UserResponse{}
	dbUsers, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	for _, dbUser := range dbUsers {
		userResponse = append(userResponse, models.UserResponse{
			ID:   dbUser.ID,
			Name: dbUser.Name,
			DOB:  dbUser.Dob,
			Age:  calculateAge(dbUser.Dob),
		})
	}
	return userResponse, nil
}

func (s *UserService) CreateUser(ctx context.Context, name string, dob time.Time) (models.UserResponse, error) {
	dbUser, err := s.repo.CreateUser(ctx, database.CreateUserParams{
		Name: name,
		Dob:  dob,
	})
	if err != nil {
		return models.UserResponse{}, err
	}
	return models.UserResponse{
		ID:   dbUser.ID,
		Name: dbUser.Name,
		DOB:  dbUser.Dob,
		Age:  calculateAge(dbUser.Dob),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int32, name string, dob time.Time) (models.UserResponse, error) {
	arg := database.UpdateUserParams{
		ID:   id,
		Name: name,
		Dob:  dob,
	}
	dbUser, err := s.repo.UpdateUser(ctx, arg)
	if err != nil {
		return models.UserResponse{}, err
	}
	return models.UserResponse{
		ID:   dbUser.ID,
		Name: dbUser.Name,
		DOB:  dbUser.Dob,
		Age:  calculateAge(dbUser.Dob),
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int32) error {
	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete user",
			zap.Int32("id", id),
			zap.Error(err),
		)
		return err
	}
	s.logger.Info("user deleted successfully", zap.Int32("id", id))
	return nil
}

func calculateAge(dob time.Time) int {
	var current time.Time = time.Now()
	var yearsApart int = current.Year() - dob.Year()
	if current.Month() < dob.Month() || (current.Month() == dob.Month() && current.Day() < dob.Day()) {
		yearsApart -= 1
	}
	age := yearsApart
	return age
}
