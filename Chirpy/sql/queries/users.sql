-- name: CreateUser :one
INSERT INTO users (id, hashed_password, created_at, updated_at, email)
VALUES (
    gen_random_uuid(),
    $1,
    NOW(),
    NOW(),
    $2
  )
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ResetUsers :exec
DELETE FROM users;