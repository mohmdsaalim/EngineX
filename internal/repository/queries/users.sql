-- name: GetUserByEmail :one
SELECT id, email, password_hash, full_name, is_active, created_at, updated_at
FROM users
WHERE email = $1 AND is_active = true;

-- name: GetUserByID :one
SELECT id, email, password_hash, full_name, is_active, created_at, updated_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, full_name)
VALUES ($1, $2, $3)
RETURNING id, email, password_hash, full_name, is_active, created_at, updated_at;