// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const connectAuthPlatform = `-- name: ConnectAuthPlatform :one
UPDATE users SET auth_providers = array_append(auth_providers, $1) WHERE id = $2 AND email = $3 AND deleted_at IS NULL AND array_position(auth_providers, $1) IS NULL
RETURNING id, auth_providers
`

type ConnectAuthPlatformParams struct {
	ArrayAppend interface{}
	ID          int64
	Email       string
}

type ConnectAuthPlatformRow struct {
	ID            int64
	AuthProviders []string
}

func (q *Queries) ConnectAuthPlatform(ctx context.Context, arg ConnectAuthPlatformParams) (ConnectAuthPlatformRow, error) {
	row := q.db.QueryRow(ctx, connectAuthPlatform, arg.ArrayAppend, arg.ID, arg.Email)
	var i ConnectAuthPlatformRow
	err := row.Scan(&i.ID, &i.AuthProviders)
	return i, err
}

const createUser = `-- name: CreateUser :exec
INSERT INTO users (
  name, email, password, auth_providers, picture, token_hash, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
`

type CreateUserParams struct {
	Name          string
	Email         string
	Password      string
	AuthProviders []string
	Picture       *string
	TokenHash     string
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
	_, err := q.db.Exec(ctx, createUser,
		arg.Name,
		arg.Email,
		arg.Password,
		arg.AuthProviders,
		arg.Picture,
		arg.TokenHash,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const getUser = `-- name: GetUser :one
SELECT id, name, email, auth_providers, picture, email_verified, created_at, updated_at FROM users WHERE id=$1 AND deleted_at IS NULL
`

type GetUserRow struct {
	ID            int64
	Name          string
	Email         string
	AuthProviders []string
	Picture       *string
	EmailVerified bool
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
}

func (q *Queries) GetUser(ctx context.Context, id int64) (GetUserRow, error) {
	row := q.db.QueryRow(ctx, getUser, id)
	var i GetUserRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.AuthProviders,
		&i.Picture,
		&i.EmailVerified,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserSecrets = `-- name: GetUserSecrets :one
SELECT id, password, token_hash, auth_providers FROM users WHERE email=$1 AND deleted_at IS NULL
`

type GetUserSecretsRow struct {
	ID            int64
	Password      string
	TokenHash     string
	AuthProviders []string
}

func (q *Queries) GetUserSecrets(ctx context.Context, email string) (GetUserSecretsRow, error) {
	row := q.db.QueryRow(ctx, getUserSecrets, email)
	var i GetUserSecretsRow
	err := row.Scan(
		&i.ID,
		&i.Password,
		&i.TokenHash,
		&i.AuthProviders,
	)
	return i, err
}

const getUserTokenHash = `-- name: GetUserTokenHash :one
SELECT token_hash FROM users WHERE id=$1 AND deleted_at IS NULL
`

func (q *Queries) GetUserTokenHash(ctx context.Context, id int64) (string, error) {
	row := q.db.QueryRow(ctx, getUserTokenHash, id)
	var token_hash string
	err := row.Scan(&token_hash)
	return token_hash, err
}

const unlinkAuthPlatform = `-- name: UnlinkAuthPlatform :exec
UPDATE users SET auth_providers = array_remove(auth_providers, $1) WHERE id = $2 AND deleted_at IS NULL
`

type UnlinkAuthPlatformParams struct {
	ArrayRemove interface{}
	ID          int64
}

func (q *Queries) UnlinkAuthPlatform(ctx context.Context, arg UnlinkAuthPlatformParams) error {
	_, err := q.db.Exec(ctx, unlinkAuthPlatform, arg.ArrayRemove, arg.ID)
	return err
}

const updatePassword = `-- name: UpdatePassword :exec
UPDATE users SET password = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL
`

type UpdatePasswordParams struct {
	ID        int64
	Password  string
	UpdatedAt pgtype.Timestamptz
}

func (q *Queries) UpdatePassword(ctx context.Context, arg UpdatePasswordParams) error {
	_, err := q.db.Exec(ctx, updatePassword, arg.ID, arg.Password, arg.UpdatedAt)
	return err
}
