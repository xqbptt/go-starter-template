package platformService

import (
	"backend/db"
	"context"
)

type AuthPlatform interface {
	AuthKey() string
	LoginGetEmail(ctx context.Context, payload string) (string, error)
	LoginExtraVerify(ctx context.Context, payload string, user db.GetUserSecretsRow) error
	LinkGetEmail(ctx context.Context, payload string) (string, error)
	LinkExtraInformation(ctx context.Context, userID int64, payload string) error
	GenerateDbUser(ctx context.Context, payload any) (*db.CreateUserParams, error)
}
