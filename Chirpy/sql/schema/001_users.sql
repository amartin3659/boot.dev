-- +goose Up
CREATE TABLE users (id UUID PRIMARY KEY, hashed_password TEXT NOT NULL, created_at TIMESTAMP NOT NULL, updated_at TIMESTAMP NOT NULL, email TEXT NOT NULL UNIQUE, is_chirpy_red BOOLEAN NOT NULL DEFAULT FALSE);

-- +goose Down
DROP TABLE users;
