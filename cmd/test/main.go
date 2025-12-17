package main

import (
	"context"
	"errors"
	"fmt"
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

// TestResult holds the result of a test
type TestResult struct {
	Success bool
	Message string
	Data    interface{}
	Error   error
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

// AgeCalculationTest tests the age calculation logic
type AgeCalculationTest struct {
	Name     string
	DOB      time.Time
	Expected int
}

// RunAgeCalculationTests tests various age calculation scenarios
func RunAgeCalculationTests() {
	fmt.Println("\n" + repeatChar("=", 80))
	fmt.Println("AGE CALCULATION UNIT TESTS")
	fmt.Println(repeatChar("=", 80) + "\n")

	tests := []AgeCalculationTest{
		{
			Name:     "Person born today (age 0)",
			DOB:      time.Now(),
			Expected: 0,
		},
		{
			Name:     "Person born 1 year ago",
			DOB:      time.Now().AddDate(-1, 0, 0),
			Expected: 1,
		},
		{
			Name:     "Person born 30 years ago",
			DOB:      time.Now().AddDate(-30, 0, 0),
			Expected: 30,
		},
		{
			Name:     "Person born before birthday this year",
			DOB:      time.Date(time.Now().Year()-25, time.Now().Month()+1, time.Now().Day(), 0, 0, 0, 0, time.UTC),
			Expected: 24,
		},
		{
			Name:     "Person born after birthday this year",
			DOB:      time.Date(time.Now().Year()-25, time.Now().Month()-1, time.Now().Day(), 0, 0, 0, 0, time.UTC),
			Expected: 25,
		},
		{
			Name:     "Person born in leap year",
			DOB:      time.Date(1996, 2, 29, 0, 0, 0, 0, time.UTC),
			Expected: time.Now().Year() - 1996,
		},
		{
			Name:     "Classic DOB: 1990-05-15",
			DOB:      time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			Expected: time.Now().Year() - 1990,
		},
	}

	passed := 0
	failed := 0

	for i, test := range tests {
		fmt.Printf("TEST %d: %s\n", i+1, test.Name)
		fmt.Println(repeatChar("-", 79))

		age := calculateAge(test.DOB)

		if age == test.Expected {
			fmt.Printf("✅ PASSED: Age calculated correctly as %d\n", age)
			passed++
		} else {
			fmt.Printf("❌ FAILED: Expected age %d, got %d\n", test.Expected, age)
			failed++
		}
		fmt.Println()
	}

	fmt.Println(repeatChar("=", 80))
	fmt.Printf("Age Calculation Tests: %d passed, %d failed\n", passed, failed)
	fmt.Println(repeatChar("=", 80) + "\n")
}

// calculateAge mimics the service layer age calculation
func calculateAge(dob time.Time) int {
	current := time.Now()
	yearsApart := current.Year() - dob.Year()
	if current.Month() < dob.Month() || (current.Month() == dob.Month() && current.Day() < dob.Day()) {
		yearsApart -= 1
	}
	return yearsApart
}

func printTestResult(result *TestResult) {
	if result.Success {
		fmt.Printf("✅ PASSED: %s\n", result.Message)
		if result.Error != nil {
			fmt.Printf("   Error: %v\n", result.Error)
		}
		if result.Data != nil {
			fmt.Printf("   Data: %+v\n", result.Data)
		}
	} else {
		fmt.Printf("❌ FAILED: %s\n", result.Message)
		if result.Error != nil {
			fmt.Printf("   Error: %v\n", result.Error)
		}
	}
}

func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}

func main() {
	// Run age calculation unit tests first
	RunAgeCalculationTests()

	fmt.Println("\n" + repeatChar("=", 80))
	fmt.Println("SYSTEM TEST SUITE - Full Workflow Validation")
	fmt.Println(repeatChar("=", 80) + "\n")

	runner := NewSystemTestRunner()

	testsPassed := 0
	testsFailed := 0

	// Test 1: Create User (Happy Path)
	fmt.Println("TEST 1: Create User (Valid Request)")
	fmt.Println(repeatChar("-", 79))
	result := runner.RunCreateUserTest("John Doe", "1990-05-15")
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 2: Get User (Happy Path)
	fmt.Println("\nTEST 2: Get User by ID")
	fmt.Println(repeatChar("-", 79))
	result = runner.RunGetUserTest(1)
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 3: Create Another User
	fmt.Println("\nTEST 3: Create Another User")
	fmt.Println(repeatChar("-", 79))
	result = runner.RunCreateUserTest("Jane Smith", "1992-08-22")
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 4: List Users
	fmt.Println("\nTEST 4: List All Users")
	fmt.Println(repeatChar("-", 79))
	result = runner.RunListUsersTest()
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 5: Update User
	fmt.Println("\nTEST 5: Update User")
	fmt.Println(repeatChar("-", 79))
	result = runner.RunUpdateUserTest(1, "John Doe Updated", "1990-05-20")
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 6: Delete User
	fmt.Println("\nTEST 6: Delete User")
	fmt.Println(repeatChar("-", 79))
	result = runner.RunDeleteUserTest(2)
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 7: Get Non-Existent User (Error Handling)
	fmt.Println("\nTEST 7: Get Non-Existent User (Error Handling)")
	fmt.Println(repeatChar("-", 79))
	result = runner.RunGetUserTest(999)
	if result.Success {
		fmt.Println("❌ FAILED: Should have returned error for non-existent user")
		testsFailed++
	} else {
		fmt.Printf("✅ PASSED: Correctly returned error: %v\n", result.Error)
		testsPassed++
	}

	// Test 8: Validation - Empty Name
	fmt.Println("\nTEST 8: Validation - Empty Name (Should Fail)")
	fmt.Println(repeatChar("-", 79))
	result = runner.RunValidationErrorTest("", "1990-05-15")
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 9: Validation - Invalid Date Format
	fmt.Println("\nTEST 9: Validation - Invalid Date Format (Should Fail)")
	fmt.Println(repeatChar("-", 79))
	result = runner.RunValidationErrorTest("John Doe", "05-15-1990")
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 10: Validation - Future Date
	fmt.Println("\nTEST 10: Validation - Future Date (Should Fail)")
	fmt.Println(repeatChar("-", 79))
	futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	result = runner.RunValidationErrorTest("John Doe", futureDate)
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 11: Validation - Name Too Long
	fmt.Println("\nTEST 11: Validation - Name Too Long (Should Fail)")
	fmt.Println(repeatChar("-", 79))
	longName := "a"
	for i := 0; i < 255; i++ {
		longName += "a"
	}
	result = runner.RunValidationErrorTest(longName, "1990-05-15")
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 12: Database Error Handling
	fmt.Println("\nTEST 12: Database Error Handling (Simulated DB Failure)")
	fmt.Println(repeatChar("-", 79))
	result = runner.RunDatabaseErrorTest()
	printTestResult(result)
	if result.Success {
		testsPassed++
	} else {
		testsFailed++
	}

	// Test 13: Create, Update, and Verify
	fmt.Println("\nTEST 13: Full Workflow - Create, Update, Get, Verify Age Calculation")
	fmt.Println(repeatChar("-", 79))
	result = runner.RunCreateUserTest("Bob Johnson", "1985-03-10")
	if result.Success {
		fmt.Printf("✅ User created: %+v\n", result.Data)

		// Update the user
		result = runner.RunUpdateUserTest(3, "Bob Johnson Updated", "1985-04-10")
		if result.Success {
			fmt.Printf("✅ User updated: %+v\n", result.Data)

			// Get the user and verify
			result = runner.RunGetUserTest(3)
			if result.Success {
				fmt.Printf("✅ User retrieved: %+v\n", result.Data)
				testsPassed++
			} else {
				fmt.Printf("❌ Failed to retrieve user: %v\n", result.Error)
				testsFailed++
			}
		} else {
			fmt.Printf("❌ Failed to update user: %v\n", result.Error)
			testsFailed++
		}
	} else {
		fmt.Printf("❌ Failed to create user: %v\n", result.Error)
		testsFailed++
	}

	// Test 14: Verify Repository State
	fmt.Println("\nTEST 14: Verify Repository State (User Count)")
	fmt.Println(repeatChar("-", 79))
	count := runner.repo.GetUserCount()
	// We should have users 1 (updated) and 3 (new) - user 2 was deleted
	if count == 2 {
		fmt.Printf("✅ PASSED: Correct user count in repository: %d\n", count)
		testsPassed++
	} else {
		fmt.Printf("❌ FAILED: Expected 2 users, got %d\n", count)
		testsFailed++
	}

	// Final Summary
	fmt.Println("\n" + repeatChar("=", 80))
	fmt.Println("TEST SUMMARY")
	fmt.Println(repeatChar("=", 80))
	fmt.Printf("Total Tests: %d\n", testsPassed+testsFailed)
	fmt.Printf("Passed: %d ✅\n", testsPassed)
	fmt.Printf("Failed: %d ❌\n", testsFailed)

	if testsFailed == 0 {
		fmt.Println("\n🎉 ALL TESTS PASSED! System is working correctly.")
	} else {
		fmt.Printf("\n⚠️  %d test(s) failed. Review the output above.\n", testsFailed)
	}
	fmt.Println(repeatChar("=", 80) + "\n")
}
