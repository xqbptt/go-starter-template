-- name: GetUser :one
SELECT id, name, email, auth_providers, picture, email_verified, created_at, updated_at FROM users WHERE id=$1 AND deleted_at IS NULL;

-- name: GetUserTokenHash :one
SELECT token_hash FROM users WHERE id=$1 AND deleted_at IS NULL;

-- name: CreateUser :exec
INSERT INTO users (
  name, email, password, auth_providers, picture, token_hash, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetUserSecrets :one
SELECT id, password, token_hash, auth_providers FROM users WHERE email=$1 AND deleted_at IS NULL;

-- name: ConnectAuthPlatform :one
UPDATE users SET auth_providers = array_append(auth_providers, $1) WHERE id = $2 AND email = $3 AND deleted_at IS NULL AND array_position(auth_providers, $1) IS NULL
RETURNING id, auth_providers;

-- name: UnlinkAuthPlatform :exec
UPDATE users SET auth_providers = array_remove(auth_providers, $1) WHERE id = $2 AND deleted_at IS NULL;

-- name: UpdatePassword :exec
UPDATE users SET password = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL;
