package test

import (
	"context"
	"errors"
	"sync"
	"time"
	database "user-api/db/sqlc"
	"user-api/internal/models"
	"user-api/internal/service"
	"user-api/internal/validator"

	"go.uber.org/zap"
)

// MockUserRepository is an in-memory mock implementation of UserRepository
type MockUserRepository struct {
	mu         sync.RWMutex
	users      map[int32]*database.User
	nextID     int32
	shouldFail bool
}

// NewMockUserRepository creates a new mock repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[int32]*database.User),
		nextID: 1,
	}
}

// GetUser retrieves a user by ID
func (m *MockUserRepository) GetUser(ctx context.Context, id int32) (database.User, error) {
	if m.shouldFail {
		return database.User{}, errors.New("mock database error")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return database.User{}, errors.New("user not found")
	}
	return *user, nil
}

// ListUsers retrieves all users
func (m *MockUserRepository) ListUsers(ctx context.Context) ([]database.User, error) {
	if m.shouldFail {
		return nil, errors.New("mock database error")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]database.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

// CreateUser creates a new user
func (m *MockUserRepository) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	if m.shouldFail {
		return database.User{}, errors.New("mock database error")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	user := database.User{
		ID:   m.nextID,
		Name: arg.Name,
		Dob:  arg.Dob,
	}
	m.users[m.nextID] = &user
	m.nextID++
	return user, nil
}

// UpdateUser updates an existing user
func (m *MockUserRepository) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	if m.shouldFail {
		return database.User{}, errors.New("mock database error")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[arg.ID]
	if !exists {
		return database.User{}, errors.New("user not found")
	}
	user.Name = arg.Name
	user.Dob = arg.Dob
	return *user, nil
}

// DeleteUser deletes a user
func (m *MockUserRepository) DeleteUser(ctx context.Context, id int32) error {
	if m.shouldFail {
		return errors.New("mock database error")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[id]; !exists {
		return errors.New("user not found")
	}
	delete(m.users, id)
	return nil
}

// SetShouldFail sets the repository to fail all operations
func (m *MockUserRepository) SetShouldFail(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = fail
}

// GetUserCount returns the number of users in the mock repository
func (m *MockUserRepository) GetUserCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.users)
}

// TestScenario represents a single test scenario
type TestScenario struct {
	Name          string
	Request       interface{}
	ExpectedError bool
	ValidateFn    func(t *TestResult) bool
}

// TestResult holds the result of a test
type TestResult struct {
	Success bool
	Message string
	Data    interface{}
	Error   error
}

// SystemTestRunner orchestrates the system tests
type SystemTestRunner struct {
	repo      *MockUserRepository
	service   *service.UserService
	validator *validator.Validator
	logger    *zap.Logger
}

// NewSystemTestRunner creates a new system test runner
func NewSystemTestRunner() *SystemTestRunner {
	logger, _ := zap.NewDevelopment()
	repo := NewMockUserRepository()
	userService := service.NewUserService(repo, logger)
	userValidator := validator.NewValidator()

	return &SystemTestRunner{
		repo:      repo,
		service:   userService,
		validator: userValidator,
		logger:    logger,
	}
}

// RunCreateUserTest tests user creation workflow
func (r *SystemTestRunner) RunCreateUserTest(name string, dob string) *TestResult {
	// Validate request
	req := models.CreateUserRequest{
		Name: name,
		DOB:  dob,
	}
	if err := r.validator.ValidateStruct(req); err != nil {
		return &TestResult{
			Success: false,
			Message: "Validation failed",
			Error:   err,
		}
	}

	// Parse DOB
	parsedDOB, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return &TestResult{
			Success: false,
			Message: "Date parsing failed",
			Error:   err,
		}
	}

	// Call service (orchestrates to repository)
	user, err := r.service.CreateUser(context.Background(), name, parsedDOB)
	if err != nil {
		return &TestResult{
			Success: false,
			Message: "Service call failed",
			Error:   err,
		}
	}

	return &TestResult{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	}
}

// RunGetUserTest tests retrieving a user
func (r *SystemTestRunner) RunGetUserTest(id int32) *TestResult {
	user, err := r.service.GetUser(context.Background(), id)
	if err != nil {
		return &TestResult{
			Success: false,
			Message: "Failed to get user",
			Error:   err,
		}
	}

	return &TestResult{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	}
}

// RunUpdateUserTest tests updating a user
func (r *SystemTestRunner) RunUpdateUserTest(id int32, name string, dob string) *TestResult {
	// Validate request
	req := models.UpdateUserRequest{
		Name: name,
		DOB:  dob,
	}
	if err := r.validator.ValidateStruct(req); err != nil {
		return &TestResult{
			Success: false,
			Message: "Validation failed",
			Error:   err,
		}
	}

	// Parse DOB
	parsedDOB, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return &TestResult{
			Success: false,
			Message: "Date parsing failed",
			Error:   err,
		}
	}

	// Call service
	user, err := r.service.UpdateUser(context.Background(), id, name, parsedDOB)
	if err != nil {
		return &TestResult{
			Success: false,
			Message: "Failed to update user",
			Error:   err,
		}
	}

	return &TestResult{
		Success: true,
		Message: "User updated successfully",
		Data:    user,
	}
}

// RunDeleteUserTest tests deleting a user
func (r *SystemTestRunner) RunDeleteUserTest(id int32) *TestResult {
	err := r.service.DeleteUser(context.Background(), id)
	if err != nil {
		return &TestResult{
			Success: false,
			Message: "Failed to delete user",
			Error:   err,
		}
	}

	return &TestResult{
		Success: true,
		Message: "User deleted successfully",
	}
}

// RunListUsersTest tests listing all users
func (r *SystemTestRunner) RunListUsersTest() *TestResult {
	users, err := r.service.ListUsers(context.Background())
	if err != nil {
		return &TestResult{
			Success: false,
			Message: "Failed to list users",
			Error:   err,
		}
	}

	return &TestResult{
		Success: true,
		Message: "Users listed successfully",
		Data:    users,
	}
}

// RunValidationErrorTest tests that validation properly rejects invalid input
func (r *SystemTestRunner) RunValidationErrorTest(name string, dob string) *TestResult {
	req := models.CreateUserRequest{
		Name: name,
		DOB:  dob,
	}
	err := r.validator.ValidateStruct(req)
	if err == nil {
		return &TestResult{
			Success: false,
			Message: "Validation should have failed but didn't",
		}
	}

	return &TestResult{
		Success: true,
		Message: "Validation correctly rejected invalid input",
		Error:   err,
	}
}

// RunDatabaseErrorTest tests error handling when repository fails
func (r *SystemTestRunner) RunDatabaseErrorTest() *TestResult {
	r.repo.SetShouldFail(true)
	defer r.repo.SetShouldFail(false)

	_, err := r.service.CreateUser(context.Background(), "Test User", time.Now().AddDate(-30, 0, 0))
	if err == nil {
		return &TestResult{
			Success: false,
			Message: "Database error should have been returned",
		}
	}

	return &TestResult{
		Success: true,
		Message: "Database error handled correctly",
		Error:   err,
	}
}
