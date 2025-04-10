package v1

import (
	"backend/api/apiUtils"
	"backend/api/middleware"
	"backend/dto"
	"backend/service"
	"backend/token"
	"backend/utils"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	CreateUser() gin.HandlerFunc
	GetUser() gin.HandlerFunc
	Login() gin.HandlerFunc
	ConnectAuthPlatform() gin.HandlerFunc
	UnlinkAuthPlatform() gin.HandlerFunc
	RefreshToken() gin.HandlerFunc
}

type userHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) UserHandler {
	return &userHandler{
		service: service,
	}
}

func (h *userHandler) CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createUserRequest dto.CreateUserRequest

		err := c.ShouldBind(&createUserRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, apiUtils.ValidatorError(err))
			return
		}

		err = h.service.CreateUser(c.Request.Context(), &createUserRequest)
		if err != nil {
			apiUtils.SendErrorResponse(c, err)
			return
		}

		c.Status(http.StatusCreated)
	}
}

func (h *userHandler) ConnectAuthPlatform() gin.HandlerFunc {
	return func(c *gin.Context) {
		var connectAuthPlatformRequest dto.ConnectAuthPlatformRequest

		userID, err := utils.ParseToInt64OrNotFound(c, "userID")
		if err != nil {
			return
		}

		err = c.ShouldBind(&connectAuthPlatformRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, apiUtils.ValidatorError(err))
			return
		}

		err = h.service.ConnectAuthPlatform(apiUtils.GetContextFromGinContext(c), userID, &connectAuthPlatformRequest)
		if err != nil {
			apiUtils.SendErrorResponse(c, err)
			return
		}

		c.Status(http.StatusCreated)
	}
}

func (h *userHandler) UnlinkAuthPlatform() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.ParseToInt64OrNotFound(c, "userID")
		if err != nil {
			return
		}

		provider := c.Param("provider")
		if provider == "" {
			c.JSON(http.StatusBadRequest, apiUtils.ValidatorError(errors.New("provider is required")))
			return
		}

		err = h.service.UnlinkAuthPlatform(apiUtils.GetContextFromGinContext(c), userID, provider)
		if err != nil {
			apiUtils.SendErrorResponse(c, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func (h *userHandler) GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.ParseToInt64OrNotFound(c, "userID")
		if err != nil {
			return
		}

		user, err := h.service.GetUser(apiUtils.GetContextFromGinContext(c), userID)
		if err != nil {
			slog.ErrorContext(c, "getting user failed", slog.Any("error", err), slog.Int64("userID", userID))
			apiUtils.SendErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func (h *userHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest dto.LoginRequest

		err := c.ShouldBind(&loginRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, apiUtils.ValidatorError(err))
			return
		}

		loginResponse, err := h.service.Login(c, &loginRequest)
		if err != nil {
			apiUtils.SendErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, loginResponse)
	}
}

func (h *userHandler) RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {

		newCtx := context.WithValue(
			c.Request.Context(),
			middleware.RefreshTokenPayloadKey,
			c.MustGet(fmt.Sprint(middleware.RefreshTokenPayloadKey)).(*token.RefreshPayload))

		loginResponse, err := h.service.GenerateAccessToken(newCtx)
		if err != nil {
			apiUtils.SendErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, loginResponse)
	}
}
