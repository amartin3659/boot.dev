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

-- name: UpdateUserByID :one
UPDATE users SET email = $1, hashed_password = $2 WHERE id = $3 RETURNING *;

-- name: UpdateUserChirpyRedByID :one
UPDATE users SET is_chirpy_red = $1 WHERE id = $2 RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;