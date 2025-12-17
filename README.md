# Go-Backend-Development-Task

A RESTful User API built in Go demonstrating clean architecture (handlers, services, repositories, models), structured logging, input validation, and test coverage.

This README explains how to set up the project locally, run the server, and execute the test suite (system tests and unit tests).

## Prerequisites
- Go 1.20+ installed and available on your PATH
- PostgreSQL (only required to run the server against a real DB; tests run without DB)
- Git (optional)

## Local setup

1. Clone the repository (if you haven't already):

```bash
git clone <repo-url>
cd Go-Backend-Development-Task
```

2. Download module dependencies:

```bash
go mod download
```

3. (Optional) Install the PostgreSQL driver if not already present (the project already declares this dependency in go.mod):

```bash
go get github.com/lib/pq
```

## Configuration

The application reads configuration from environment variables. Sensible defaults are provided for local development.

- `DATABASE_URL` — PostgreSQL connection string. Default: `postgres://user:password@localhost:5432/userdb?sslmode=disable`
- `PORT` — port the server listens on. Default: `8080`
- `APP_ENV` — `development` or `production` (affects logger formatting)

If you want to run the server against a real Postgres instance, create a role and database or change `DATABASE_URL` to match an existing user. Example (run as postgres superuser):

```powershell
$env:PGPASSWORD = "POSTGRES_SUPERUSER_PASSWORD"
psql -U postgres -h localhost -c "CREATE ROLE \"user\" WITH LOGIN PASSWORD 'password';"
psql -U postgres -h localhost -c "CREATE DATABASE userdb OWNER \"user\";"

# Set DATABASE_URL for current PowerShell session
$env:DATABASE_URL = "postgres://user:password@localhost:5432/userdb?sslmode=disable"

# Persist across sessions (optional)
setx DATABASE_URL "postgres://user:password@localhost:5432/userdb?sslmode=disable"
```

Adjust credentials to suit your environment.

## Running the server

Start the server locally (will try to connect to `DATABASE_URL`):

```powershell
go run cmd/server/main.go
```

If the database is unavailable, the server will fail to start. You can run the test suite (below) which uses an in-memory mock repository and does not require Postgres.

## Tests

There are two main test artifacts included:

- Age calculation unit tests — a focused set of unit tests that validate birthday and leap-year edge cases.
- System test suite — a comprehensive set of tests that exercise the full workflow (handlers → services → repository) using an in-memory mock repository (no DB required).

Both are bundled in a single test runner for convenience. Run all tests with:

```powershell
go run cmd/test/main.go
```

What it runs:

- First: age calculation unit tests (`RunAgeCalculationTests()`)
- Then: system tests covering Create, Read, Update, Delete, List, validation, and simulated DB errors

All tests are self-contained and will report a summary at the end.

## Validation

Input validation is implemented using `github.com/go-playground/validator/v10`. Key rules:

- `name`: required, 1–255 characters
- `dob`: required, must be `YYYY-MM-DD`, cannot be in the future

Validation runs in the handler layer and returns `400 Bad Request` with descriptive messages when a request fails validation.

## Project structure (high level)

- `cmd/server` — server entrypoint
- `cmd/test` — test runner (system tests + age unit tests)
- `internal/handler` — HTTP handlers and request parsing/validation
- `internal/service` — business logic (age calculation, orchestration)
- `internal/repository` — repository interfaces and adapter implementations
- `internal/models` — API request/response models
- `internal/validator` — validation helpers and custom rules
- `db/sqlc` — sqlc-generated DB code (if using Postgres)

---

Original submission: internship task for Ainyx solutions.
