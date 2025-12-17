# Project Reasoning & Design Decisions

## Overview
This project is a RESTful User API built with Go, implementing a clean architecture pattern with proper separation of concerns. The API provides CRUD operations for user management with automatic age calculation based on date of birth.

## Architecture Decisions

### 1. Layered Architecture
I chose a layered architecture with clear separation between handler, service, repository, and models layers. This decision was driven by:

**Handler Layer** (`internal/handler/`): Handles HTTP request/response logic, input validation, and status codes. Keeps web framework concerns isolated from business logic.

**Service Layer** (`internal/service/`): Contains business logic including age calculation. This layer orchestrates between handlers and repositories, making it easy to add features like caching or additional validation without touching HTTP or database code.

**Repository Layer** (`internal/repository/`): Abstracts database operations through interfaces. This makes the code testable (can mock repositories) and allows switching databases without changing business logic.

**Models Layer** (`internal/models/`): Separates API request/response models from database models. The API uses `string` for DOB in requests (easier for clients) while internally we work with `time.Time`, preventing tight coupling between API contracts and internal representations.

### 2. Dependency Injection
I used constructor functions (`NewUserHandler`, `NewUserService`, etc.) throughout the codebase to inject dependencies. This approach:
- Makes dependencies explicit and visible
- Enables easy unit testing through mocking
- Follows Go idioms for component initialization
- Allows flexible composition in `main.go`

### 3. Interface-Based Repository Pattern
The `UserRepository` interface in `repository/user_repository.go` provides abstraction over the concrete implementation. Benefits:
- The service layer depends on abstractions, not concrete implementations (Dependency Inversion Principle)
- Easy to write tests by creating mock implementations
- Could swap out sqlc for another ORM or raw SQL without changing service layer
- Makes the codebase more maintainable and extensible

### 4. Structured Logging with Zap
I implemented structured logging using uber's zap library instead of standard `log` package because:
- **Performance**: Zap is significantly faster with zero-allocation logging
- **Structure**: Logs are structured as key-value pairs, making them machine-readable and easy to query
- **Environment-aware**: Different configurations for development (colored, readable) vs production (JSON, timestamp formatting)
- **Context**: Each log entry includes relevant context (HTTP method, path, duration, status codes)

The logging strategy includes:
- Request logging middleware tracking all HTTP requests
- Error logging for debugging issues
- Info logs for successful operations
- Centralized logger initialization from environment variables

### 5. Middleware Chain
I created custom middleware for cross-cutting concerns:

**RequestLogger**: Logs every HTTP request with method, path, status, duration, IP, and user agent. This provides observability and helps debug issues in production.

**ErrorHandler**: Centralized error handling that catches unhandled errors, logs them with context, and returns consistent error responses to clients.

**CORS**: Handles Cross-Origin Resource Sharing to allow browser-based clients to access the API from different origins.

This middleware approach keeps handlers clean and focused on business logic while ensuring consistent behavior across all endpoints.

### 6. Database Layer with sqlc
I used sqlc to generate type-safe Go code from SQL queries rather than using an ORM. This choice provides:
- **Type Safety**: Compile-time errors if SQL and Go types don't match
- **Performance**: Near-raw SQL performance, no ORM overhead
- **Explicitness**: SQL queries are visible and reviewable in the codebase
- **Migration-friendly**: SQL files serve as documentation and migration source

The generated code in `db/sqlc/` provides a clean interface that the repository layer wraps, keeping database concerns separate from business logic.

### 7. Error Handling Strategy
I implemented defensive error handling throughout:
- Parse errors return `400 Bad Request` with descriptive messages
- Database errors return `500 Internal Server Error` after logging
- Not found errors return `404 Not Found`
- All errors are logged with context for debugging

The error responses use a consistent `fiber.Map{"error": "message"}` format, making it easy for API consumers to handle errors programmatically.

### 8. Input Validation with go-playground/validator
I implemented robust input validation using the `go-playground/validator` library (`internal/validator/`). This decision provides:
- **Declarative Validation**: Validation rules are defined as struct tags, making them visible alongside model definitions
- **Custom Rules**: Created custom validators for date format validation and future date prevention
- **User-Friendly Errors**: Validation errors return descriptive messages like "DOB must be in YYYY-MM-DD format" instead of generic errors
- **Separation of Concerns**: Validation is isolated in its own layer, keeping handlers clean
- **Consistency**: All request validation follows the same pattern and rules

The validator integration:
- Validates request DTOs (`CreateUserRequest`, `UpdateUserRequest`) in the handler layer
- Returns `400 Bad Request` with detailed error messages when validation fails
- Logs validation failures for debugging
- Prevents invalid data from reaching the service layer

Validation rules enforced:
- Name: required, 1-255 characters
- DOB: required, YYYY-MM-DD format, cannot be a future date

### 8. Date Handling and Age Calculation
Dates are handled as strings in API requests (`"2006-01-02"` format) but stored and processed as `time.Time` internally. The age calculation function:
- Accurately handles birthday edge cases (checks month and day)
- Is pure and testable (no side effects)
- Returns age as of the current date

This separation between API format (string) and internal format (time.Time) gives flexibility to change date formatting without affecting internal logic.

### 9. Configuration Management
Configuration is loaded from environment variables with sensible defaults:
- `DATABASE_URL`: Connection string (defaults to local development DB)
- `PORT`: Server port (defaults to 8080)
- `APP_ENV`: Environment mode affecting log format (defaults to "development")

This follows the [12-factor app](https://12factor.net/) methodology, making the application easy to deploy across different environments without code changes.

### 10. Graceful Shutdown
Implemented graceful shutdown handling that:
- Listens for SIGINT/SIGTERM signals
- Allows in-flight requests to complete
- Closes database connections properly
- Logs the shutdown process

This prevents data corruption and ensures clean deployments.

## Key Design Patterns Used

1. **Repository Pattern**: Abstracts data access
2. **Dependency Injection**: Through constructors
3. **Factory Pattern**: Constructor functions create configured instances
4. **Middleware Pattern**: Composable request processing
5. **Interface Segregation**: Small, focused interfaces

## Trade-offs and Considerations

### What I Prioritized
- **Code clarity over cleverness**: Straightforward implementations that are easy to understand
- **Separation of concerns**: Each layer has a single, well-defined responsibility
- **Type safety**: Leveraging Go's type system to catch errors at compile time
- **Observability**: Comprehensive logging for debugging and monitoring

### What Could Be Improved with More Time
- **API documentation**: Could add OpenAPI/Swagger documentation
- **Rate limiting**: Could add rate limiting middleware to prevent abuse
- **Pagination**: List endpoint could benefit from pagination for large datasets
- **Database migrations**: Could add a migration tool like golang-migrate for version control

## Testing Strategy

### System Tests
Created a comprehensive system test suite (`cmd/test/main.go`) that validates the entire workflow without requiring a database connection. The test suite includes:

- **CRUD Operations**: Tests for Create, Read (Get), Update, Delete, and List operations
- **Validation Integration**: Tests that validation rules properly reject invalid input (empty names, invalid date formats, future dates, names exceeding 255 characters)
- **Error Handling**: Tests for non-existent user retrieval, database failure simulation, and error propagation through layers
- **Full Workflows**: End-to-end tests combining multiple operations (create → update → get → delete)
- **Repository State Verification**: Confirms the in-memory repository maintains correct state

**Key Innovation**: Built a `MockUserRepository` that implements the `UserRepository` interface without touching the database, allowing tests to run independently of PostgreSQL availability.

### Age Calculation Unit Tests
Dedicated unit tests (`RunAgeCalculationTests()` in `cmd/test/main.go`) verify the age calculation logic handles edge cases correctly:

- **Birthday Logic**: Tests that age increments only after the birthday has passed in the current year
- **Edge Cases**: 
  - Person born today (age 0)
  - Person born 1 year ago (age 1)
  - Person born 30 years ago
  - Person born before/after birthday this year
  - Person born in leap year (Feb 29)
  - Classic test case (1990-05-15)

All tests pass, confirming the age calculation is accurate and handles edge cases properly.

### Running Tests
Execute with:
```bash
go run cmd\test\main.go
```

This runs all age calculation unit tests followed by the complete system test suite (14 tests total, all passing).

## Why This Architecture?

Coming from other programming languages, I wanted to build something that follows Go best practices while remaining maintainable. The layered architecture is a proven pattern that:
- Makes the codebase easy to navigate
- Allows team members to work on different layers without conflicts
- Provides natural boundaries for testing
- Scales well as the application grows

The use of interfaces and dependency injection might seem like overkill for a small API, but these patterns pay dividends in larger codebases and demonstrate understanding of SOLID principles, which I believe is valuable even for an internship.

## What I Learned

Building this project helped me understand:
- Go's approach to error handling (explicit vs exceptions)
- How interfaces enable testability and flexibility
- The power of code generation tools like sqlc
- Go's concurrency primitives for graceful shutdown
- The Fiber framework's middleware system
- Structured logging benefits over simple print statements

## How It Works End-to-End

1. **Startup**: `main.go` initializes logger, connects to database, wires up dependencies, and starts server
2. **Request**: Client sends HTTP request to `/api/v1/users`
3. **Middleware**: Request passes through CORS → ErrorHandler → RequestLogger
4. **Handler**: `UserHandler` parses request, validates input
5. **Service**: `UserService` contains business logic, calls repository
6. **Repository**: `UserRepository` executes SQL via sqlc-generated code
7. **Database**: PostgreSQL performs query and returns results
8. **Response**: Data flows back through layers, transformed to API response format
9. **Logging**: Request details logged with status, duration, and any errors

The entire flow maintains separation of concerns, with each layer having a specific job and depending only on abstractions.

---

This architecture demonstrates my understanding of backend development principles while being pragmatic enough for a small API. I'm excited to discuss any aspect of these decisions in detail during the interview.