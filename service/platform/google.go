package platformService

import (
	"backend/api/apiUtils"
	"backend/db"
	"backend/dto"
	"backend/external-api/platform/google"
	"backend/utils"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type googleService struct {
	pool   *pgxpool.Pool
	config utils.GoogleConfig
}

func NewGoogleService(pool *pgxpool.Pool, config utils.GoogleConfig) *googleService {
	return &googleService{
		pool:   pool,
		config: config,
	}
}

func (s *googleService) AuthKey() string {
	return "google"
}

func (s *googleService) LoginGetEmail(ctx context.Context, payload string) (string, error) {
	tokenResponse, err := google.GetAccessToken(payload, s.config)
	if err != nil {
		return "", err
	}

	userInfo, err := google.UserInformation(tokenResponse.AccessToken, s.config)
	if err != nil {
		return "", err
	}

	return userInfo.Email, nil
}

func (s *googleService) LoginExtraVerify(ctx context.Context, payload string, user db.GetUserSecretsRow) error {
	return nil
}

func (s *googleService) LinkExtraInformation(ctx context.Context, userID int64, payload string) error {
	return nil
}

func (s *googleService) LinkGetEmail(ctx context.Context, payload string) (string, error) {
	return s.LoginGetEmail(ctx, payload)
}

func (s *googleService) GenerateDbUser(ctx context.Context, p any) (*db.CreateUserParams, error) {
	var payload dto.CreateUserPayloadGoogle
	err := apiUtils.AssignAndValidateCreateUserPayload(ctx, p, &payload)
	if err != nil {
		return nil, err
	}

	tokenResponse, err := google.GetAccessToken(payload.AuthorizationCode, s.config)
	if err != nil {
		return nil, err
	}

	userInfo, err := google.UserInformation(tokenResponse.AccessToken, s.config)
	if err != nil {
		return nil, err
	}

	return &db.CreateUserParams{
		Name:      userInfo.Name,
		Email:     userInfo.Email,
		Password:  "NA",
		Picture:   nil,
		TokenHash: utils.GenerateRandomString(15),
	}, nil
}
