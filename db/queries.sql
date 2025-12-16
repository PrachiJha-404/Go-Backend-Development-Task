-- name: CreateUser :one
INSERT INTO users (name, dob)
VALUES ($1, $2)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id=$1 LIMIT 1;

-- name: DeleteUser :one
DELETE FROM users
WHERE id=$1
RETURNING *;

-- name: ListUsers :many
SELECT * FROM users;

-- name: UpdateUser :one
UPDATE users
SET name=$2,
dob=$3
WHERE id = $1
RETURNING *;