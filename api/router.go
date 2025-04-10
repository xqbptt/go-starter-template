package api

import (
	"backend/api/middleware"
	v1 "backend/api/v1"
	"backend/service"
	"backend/token"
	"backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterPath(
	r *gin.RouterGroup,
	config utils.Config,
	serviceName string,
	version string,
	tokenMaker token.Maker,
	userService service.UserService,
) {

	v1Route := r.Group("/v1")

	// handlers
	userHandler := v1.NewUserHandler(userService)

	// user
	userRouter := v1Route.Group("/users")
	userRouter.POST("/", userHandler.CreateUser())
	userRouter.GET("/:userID", middleware.AuthMiddleware(tokenMaker), userHandler.GetUser())
	userRouter.POST("/token", userHandler.Login())
	userRouter.POST("/refresh-token", middleware.RefreshTokenValidateMiddleware(tokenMaker), userHandler.RefreshToken())
	userRouter.POST("/:userID/auth", middleware.AuthMiddleware(tokenMaker), userHandler.ConnectAuthPlatform())
	userRouter.DELETE("/:userID/auth/:provider", middleware.AuthMiddleware(tokenMaker), userHandler.UnlinkAuthPlatform())

	// health check
	r.GET("/health", func(c *gin.Context) {
		response := map[string]string{"status": "UP", "service": serviceName, "version": version}
		c.JSON(http.StatusOK, response)
	})
}
