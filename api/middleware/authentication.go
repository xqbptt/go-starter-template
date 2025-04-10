package middleware

import (
	"backend/dto"
	"backend/token"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type key int

const (
	AuthenticationPayloadKey key = iota
)

const (
	authenticationHeaderKey  = "authorization"
	authenticationTypeBearer = "bearer"
)

var (
	ErrHeaderNotProvided = errors.New("authentication header is not provided")
)

func getPayloadFromContext(c *gin.Context, tokenMaker token.Maker) (*token.Payload, *gin.Context, error) {
	authenticationHeader := c.GetHeader(authenticationHeaderKey)
	if len(authenticationHeader) == 0 {
		return nil, c, ErrHeaderNotProvided
	}

	fields := strings.Fields(authenticationHeader)
	if len(fields) < 2 {
		err := errors.New("invalid authentication header format")
		return nil, c, err
	}

	authenticationType := strings.ToLower(fields[0])
	if authenticationType != authenticationTypeBearer {
		err := fmt.Errorf("unsupported authentication type %s", authenticationType)
		return nil, c, err
	}

	accessToken := fields[1]
	payload, err := tokenMaker.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, c, err
	}

	c.Set(fmt.Sprint(AuthenticationPayloadKey), payload)
	// ctx := utils.AppendCtx(c, slog.Int64("user_id", payload.UserID)).(*gin.Context) // TODO: check on how to make this work?

	slog.InfoContext(c, "authenticated user token",
		slog.String("token_id", payload.ID.String()),
		slog.String("token_type", payload.Type),
		slog.String("issuer", payload.Issuer),
		slog.Int64("user_id", payload.UserID),
	)

	return payload, c, nil
}

// AuthMiddleware creates a gin middleware for authentication
func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ctx, err := getPayloadFromContext(c, tokenMaker)
		if err != nil {
			slog.InfoContext(ctx, "token validation failed", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewError(err.Error()))
		}

		ctx.Next()
	}
}

// Do authentication if token exists, if not skip token validation
func AuthMiddlewareOptional(tokenMaker token.Maker) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ctx, err := getPayloadFromContext(c, tokenMaker)
		if err != nil && err != ErrHeaderNotProvided {
			slog.InfoContext(ctx, "token validation failed (optional header)", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewError(err.Error()))
		}

		ctx.Next()
	}
}
