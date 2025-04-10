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

const (
	RefreshTokenPayloadKey key = iota
)

const (
	refreshHeaderKey  = "authorization"
	refreshTypeBearer = "bearer"
)

func validateAndSetRefreshTokenContext(tokenMaker token.Maker, ctx *gin.Context) (*token.RefreshPayload, *gin.Context, error) {
	refreshHeader := ctx.GetHeader(refreshHeaderKey)

	if len(refreshHeader) == 0 {
		err := errors.New("authentication header is not provided")
		return nil, ctx, err
	}

	fields := strings.Fields(refreshHeader)
	if len(fields) < 2 {
		err := errors.New("invalid authentication header format")
		return nil, ctx, err
	}

	authenticationType := strings.ToLower(fields[0])
	if authenticationType != refreshTypeBearer {
		err := fmt.Errorf("unsupported authentication type %s", authenticationType)
		return nil, ctx, err
	}

	refreshToken := fields[1]
	payload, err := tokenMaker.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ctx, err
	}

	ctx.Set(fmt.Sprint(AuthenticationPayloadKey), payload)

	slog.InfoContext(ctx, "authenticated user token",
		slog.String("token_id", payload.ID.String()),
		slog.String("token_type", payload.Type),
		slog.String("issuer", payload.Issuer),
		slog.Int64("user_id", payload.UserID),
	)

	return payload, ctx, nil
}

func RefreshTokenValidateMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ctx, err := validateAndSetRefreshTokenContext(tokenMaker, c)
		if err != nil {
			slog.InfoContext(ctx, "refresh-token validation failed", slog.Any("error", err))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewError(err.Error()))
		}

		ctx.Next()
	}
}
