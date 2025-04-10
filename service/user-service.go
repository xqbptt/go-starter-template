package service

import (
	"backend/api/apiUtils"
	"backend/api/middleware"
	"backend/db"
	"backend/dto"
	platformService "backend/service/platform"
	"backend/token"
	"backend/utils"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService interface {
	CreateUser(context.Context, *dto.CreateUserRequest) error
	GetUser(context.Context, int64) (*dto.GetUserResponse, error)
	Login(context.Context, *dto.LoginRequest) (*dto.LoginResponse, error)
	ConnectAuthPlatform(context.Context, int64, *dto.ConnectAuthPlatformRequest) error
	UnlinkAuthPlatform(context.Context, int64, string) error
	GenerateAccessToken(context.Context) (*dto.LoginResponse, error)
}

type userService struct {
	tokenMaker    token.Maker
	pool          *pgxpool.Pool
	authPlatforms []platformService.AuthPlatform
}

func NewUserService(pool *pgxpool.Pool, tokenMaker token.Maker, authPlatforms []platformService.AuthPlatform) UserService {
	service := &userService{
		pool:          pool,
		tokenMaker:    tokenMaker,
		authPlatforms: authPlatforms,
	}

	// current platform in itself an auth platform
	service.authPlatforms = append(service.authPlatforms, service)

	return service
}

func (s *userService) CreateUser(ctx context.Context, request *dto.CreateUserRequest) error {
	slog.InfoContext(ctx, "creating user",
		slog.String("provider", request.Provider),
		slog.Any("payload", request.Payload),
	)

	for _, provider := range s.authPlatforms {
		if provider.AuthKey() == request.Provider {
			dbUser, err := provider.GenerateDbUser(ctx, request.Payload)
			if err != nil {
				return err
			}
			return s.CreateDbUser(ctx, dbUser, provider.AuthKey())
		}
	}
	return dto.NewErrorWithStatus(http.StatusBadRequest, "invalid provider")
}

func (s *userService) CreateDbUser(ctx context.Context, request *db.CreateUserParams, authProvider string) error {
	userCreateParams := db.CreateUserParams{
		Name:          request.Name,
		Email:         request.Email,
		Password:      request.Password,
		Picture:       request.Picture,
		AuthProviders: []string{authProvider},
		TokenHash:     utils.GenerateRandomString(15),
		CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		slog.Error("could not establish db connection")
		return err
	}

	defer conn.Release()
	repo := db.New(conn)
	err = repo.CreateUser(ctx, userCreateParams)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "unique_email" { // TODO: postgres specific code
				slog.ErrorContext(ctx, "email is not unique", slog.String("email", request.Email))
				return dto.NewErrorWithStatus(http.StatusBadRequest, "email already exists")
			}
		}
		slog.ErrorContext(ctx, "could not create user", slog.String("email", request.Email), slog.Any("error", err))
		return dto.NewError("could not create user")
	}

	return nil
}

func (s *userService) ConnectAuthPlatform(ctx context.Context, userID int64, request *dto.ConnectAuthPlatformRequest) error {
	currentUser := ctx.Value(middleware.AuthenticationPayloadKey).(*token.Payload)

	if currentUser.UserID != userID {
		return dto.NewErrorWithStatus(http.StatusForbidden, "user not found")
	}

	slog.InfoContext(ctx, "adding auth platform to user",
		slog.String("provider", request.Provider),
		slog.Int64("userID", currentUser.UserID),
	)

	for _, provider := range s.authPlatforms {
		if provider.AuthKey() == request.Provider {
			return s.validateAndConnectAuthPlatform(ctx, userID, provider, request.Payload)
		}
	}

	return dto.NewErrorWithStatus(http.StatusBadRequest, fmt.Sprintf("provider %s is invalid", request.Provider))
}

func (s *userService) UnlinkAuthPlatform(ctx context.Context, userID int64, provider string) error {
	currentUser := ctx.Value(middleware.AuthenticationPayloadKey).(*token.Payload)

	if currentUser.UserID != userID {
		return dto.NewErrorWithStatus(http.StatusForbidden, "user not found")
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		slog.Error("could not establish db connection")
		return err
	}

	defer conn.Release()
	repo := db.New(conn)

	user, err := repo.GetUser(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "could not get user", slog.Any("error", err))
		return dto.NewError("could not get user")
	}

	if !utils.SliceContains(user.AuthProviders, provider) {
		return nil
	}

	if len(user.AuthProviders) == 1 {
		return dto.NewErrorWithStatus(http.StatusBadRequest, "cannot unlink the last auth provider")
	}

	err = repo.UnlinkAuthPlatform(ctx, db.UnlinkAuthPlatformParams{
		ID:          userID,
		ArrayRemove: provider,
	})
	if err != nil {
		slog.ErrorContext(ctx, "could not unlink auth platform", slog.Any("error", err))
		return dto.NewError("could not unlink auth platform")
	}

	return nil
}

func (s *userService) validateAndConnectAuthPlatform(ctx context.Context, userID int64, provider platformService.AuthPlatform, payload string) error {
	email, err := provider.LinkGetEmail(ctx, payload)
	if err != nil {
		slog.ErrorContext(ctx, "could not get email", slog.Any("error", err))
		return dto.NewError("could not get email")
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		slog.Error("could not establish db connection")
		return err
	}
	defer conn.Release()

	err = provider.LinkExtraInformation(ctx, userID, payload)
	if err != nil {
		slog.ErrorContext(ctx, "could not link extra information", slog.Any("error", err))
		return err
	}

	repo := db.New(conn)
	connectedAccount, err := repo.ConnectAuthPlatform(ctx, db.ConnectAuthPlatformParams{
		ID:          userID,
		Email:       email,
		ArrayAppend: provider.AuthKey(),
	})
	slog.InfoContext(ctx, "connected account", slog.Any("connectedAccount", connectedAccount))
	if err != nil && !utils.SliceContains(connectedAccount.AuthProviders, provider.AuthKey()) {
		slog.ErrorContext(ctx, "could not connect auth platform", slog.Any("error", err))
		return dto.NewError("could not connect auth platform")
	}

	return nil
}

func (s *userService) GetUser(ctx context.Context, userID int64) (*dto.GetUserResponse, error) {
	currentUser := ctx.Value(middleware.AuthenticationPayloadKey).(*token.Payload)

	slog.InfoContext(ctx, "get the user details", slog.Int64("loggedInUserID", currentUser.UserID), slog.Int64("searchedUserID", userID))

	if userID != currentUser.UserID {
		return nil, dto.NewErrorWithStatus(http.StatusNotFound, "user not found")
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		slog.Error("could not establish db connection")
		return nil, err
	}

	defer conn.Release()
	repo := db.New(conn)
	user, err := repo.GetUser(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.ErrorContext(ctx, "user not found in db", slog.Int64("searchedUserID", userID))
			return nil, dto.NewErrorWithStatus(http.StatusNotFound, "user not found")
		}

		slog.ErrorContext(ctx, "could not get user", slog.Any("error", err))
		return nil, dto.NewError("could not get user")
	}

	slog.InfoContext(ctx, "got the user", slog.Int64("searchedUserID", userID), slog.String("email", user.Email))
	return dto.GetUserResponseFromDB(&user), nil
}

func (s *userService) Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error) {
	slog.InfoContext(ctx, "logging in user",
		slog.String("provider", request.Provider),
	)

	for _, provider := range s.authPlatforms {
		if provider.AuthKey() == request.Provider {
			return s.LoginWithProvider(ctx, provider, request.Payload)
		}
	}

	return nil, dto.NewErrorWithStatus(http.StatusBadRequest, "invalid provider")
}

func (s *userService) LoginWithProvider(ctx context.Context, provider platformService.AuthPlatform, payload string) (*dto.LoginResponse, error) {
	email, err := provider.LoginGetEmail(ctx, payload)
	if err != nil {
		slog.ErrorContext(ctx, "could not get email", slog.Any("error", err))
		return nil, dto.NewError("could not get email")
	}
	slog.Info("got email for logging in user", slog.String("email", email))

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		slog.Error("could not establish db connection")
		return nil, err
	}

	defer conn.Release()
	repo := db.New(conn)
	user, err := repo.GetUserSecrets(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.ErrorContext(ctx, "email not in found", slog.String("email", email))
			return nil, dto.NewErrorWithStatus(http.StatusForbidden, "invalid credendials")
		}
		slog.ErrorContext(ctx, "could not get user secrets", slog.Any("error", err))
		return nil, dto.NewError("could not get user")
	}

	if !utils.SliceContains(user.AuthProviders, provider.AuthKey()) {
		slog.ErrorContext(ctx, "user not found", slog.String("email", email))
		return nil, dto.NewErrorWithStatus(http.StatusForbidden, fmt.Sprintf("account not found for %s, login to account then link %s", provider.AuthKey(), provider.AuthKey()))
	}

	if err = provider.LoginExtraVerify(ctx, payload, user); err != nil {
		slog.ErrorContext(ctx, "could not verify user", slog.Any("error", err))
		return nil, dto.NewErrorWithStatus(http.StatusForbidden, "invalid credendials")
	}

	slog.InfoContext(ctx, "generating tokens for user", slog.String("email", email), slog.Int64("userID", user.ID))
	return s.generateTokens(ctx, user.ID, user.TokenHash)
}

func (s *userService) AuthKey() string {
	return "normal"
}

func (s *userService) LoginGetEmail(c context.Context, payload string) (string, error) {
	parts := strings.SplitN(payload, "|", 2)
	if len(parts) != 2 {
		slog.Error("payload is not valid", slog.String("payload", payload))
		return "", dto.NewErrorWithStatus(http.StatusBadRequest, "invalid payload")
	}

	return parts[0], nil
}

func (s *userService) LinkGetEmail(c context.Context, payload string) (string, error) {
	currentUser := c.Value(middleware.AuthenticationPayloadKey).(*token.Payload)

	conn, err := s.pool.Acquire(c)
	if err != nil {
		slog.Error("could not establish db connection")
		return "", err
	}
	defer conn.Release()

	repo := db.New(conn)
	user, err := repo.GetUser(c, currentUser.UserID)
	if err != nil {
		slog.ErrorContext(c, "error while getting user", slog.Any("error", err))
		return "", dto.NewError("could not get user details to generate token")
	}

	return user.Email, nil
}

func (s *userService) LoginExtraVerify(ctx context.Context, payload string, user db.GetUserSecretsRow) error {
	parts := strings.SplitN(payload, "|", 2)
	err := utils.CheckPassword(parts[1], user.Password)
	if err != nil {
		slog.ErrorContext(ctx, "password is incorrect", slog.String("email", parts[0]), slog.Any("error", err))
		return dto.NewErrorWithStatus(http.StatusForbidden, "invalid credendials")
	}

	return nil
}

func (s *userService) LinkExtraInformation(ctx context.Context, userID int64, payload string) error {
	if len(payload) < 8 {
		return dto.NewErrorWithStatus(http.StatusBadRequest, "password should be at least 8 characters long")
	}

	hashedPassword, err := utils.HashPassword(payload)
	if err != nil {
		slog.ErrorContext(ctx, "could not hash password", slog.Int64("userId", userID))
		return dto.NewError("could not hash password")
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		slog.Error("could not establish db connection")
		return err
	}

	defer conn.Release()
	repo := db.New(conn)
	err = repo.UpdatePassword(ctx, db.UpdatePasswordParams{
		ID:        userID,
		Password:  string(hashedPassword),
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		slog.ErrorContext(ctx, "could not update password", slog.Int64("userId", userID), slog.Any("error", err))
		return dto.NewError("could not update password")
	}

	return nil
}

func (s *userService) GenerateDbUser(ctx context.Context, p any) (*db.CreateUserParams, error) {
	var payload dto.CreateUserPayloadNormal
	err := apiUtils.AssignAndValidateCreateUserPayload(ctx, p, &payload)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		slog.ErrorContext(ctx, "could not hash password", slog.String("email", payload.Email))
		return nil, dto.NewError("could not hash password")
	}

	return &db.CreateUserParams{
		Name:      payload.Name,
		Email:     payload.Email,
		Password:  string(hashedPassword),
		Picture:   payload.Picture,
		TokenHash: utils.GenerateRandomString(15),
	}, nil
}

func (s *userService) GenerateAccessToken(c context.Context) (*dto.LoginResponse, error) {
	refreshPayload := c.Value(middleware.RefreshTokenPayloadKey).(*token.RefreshPayload)
	ctx := utils.AppendCtx(c, slog.Int64("user_id", refreshPayload.UserID))

	slog.InfoContext(ctx, "generating access token from refresh token")

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		slog.Error("could not establish db connection")
		return nil, err
	}

	defer conn.Release()
	repo := db.New(conn)
	tokenHash, err := repo.GetUserTokenHash(ctx, refreshPayload.UserID)
	if err != nil {
		slog.ErrorContext(ctx, "error while getting user", slog.Any("error", err))
		return nil, dto.NewError("could not get user details to generate token")
	}

	// check token hash
	err = s.tokenMaker.ValidateRefreshHash(refreshPayload.Hash, refreshPayload.UserID, tokenHash)
	if err != nil {
		slog.ErrorContext(ctx, "refresh token hash mismatch", slog.Any("error", err))
		return nil, dto.NewErrorWithStatus(http.StatusForbidden, err.Error())
	}

	return s.generateTokens(ctx, refreshPayload.UserID, tokenHash)
}

func (s *userService) generateTokens(ctx context.Context, userID int64, tokenHash string) (*dto.LoginResponse, error) {
	var response dto.LoginResponse
	var err error

	response.AccessToken, err = s.tokenMaker.CreateAccessToken(userID)
	if err != nil {
		slog.ErrorContext(ctx, "could not access create token", slog.Any("error", err))
		return nil, dto.NewError("could not access create token")
	}

	response.RefreshToken, err = s.tokenMaker.CreateRefreshToken(userID, tokenHash)
	if err != nil {
		slog.ErrorContext(ctx, "could not refresh create token", slog.Any("error", err))
		return nil, dto.NewError("could not refresh create token")
	}

	return &response, nil
}
