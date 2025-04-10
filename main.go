package main

import (
	"backend/api"
	"backend/api/apiUtils"
	"backend/api/middleware"
	"backend/db"
	"backend/service"
	platformService "backend/service/platform"
	"backend/token"
	"backend/utils"
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	SERVICE_NAME    = "backend-service"
	CURRENT_VERSION = "0.1.0"
)

func main() {
	ctx := context.Background()
	logger := slog.New(&utils.ContextHandler{Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})})
	slog.SetDefault(logger)

	config, err := utils.LoadConfig("config")
	if err != nil {
		slog.Error("cannot load config", slog.Any("error", err))
		os.Exit(1)
	}

	// storage, err := oci.NewOciStorage(config.OCI_STORAGE)
	if err != nil {
		slog.Error("cannot create connection to object storage", slog.Any("error", err))
		os.Exit(1)
	}

	pool, err := db.Connect(ctx, config.DB_URL)
	if err != nil {
		slog.Error("cannot connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	tokenMaker, err := token.NewJWTMaker(
		"backend.user",
		config.TOKEN.ACCESS_SECRET_KEY,
		config.TOKEN.ACCESS_PUBLIC_KEY,
		config.TOKEN.REFRESH_SECRET_KEY,
		config.TOKEN.REFRESH_PUBLIC_KEY,
		config.TOKEN.ACCESS_TOKEN_DURATION,
		config.TOKEN.REFRESH_TOKEN_DURATION,
	)
	if err != nil {
		slog.Error("cannot create token maker", slog.Any("error", err))
		os.Exit(1)
	}

	// services
	googleService := platformService.NewGoogleService(pool, config.GOOGLE)

	userService := service.NewUserService(pool, tokenMaker, []platformService.AuthPlatform{googleService})

	err = apiUtils.AddCustomValidator(binding.Validator.Engine())
	if err != nil {
		slog.Error("could not add validator", slog.Any("error", err))
		os.Exit(1)
	}

	r := gin.Default()
	r.Use(middleware.CORSMiddleware(config.CORS))
	// r.Use(func(ctx *gin.Context) { time.Sleep(500 * time.Millisecond); ctx.Next() })
	api.RegisterPath(&r.RouterGroup, config, SERVICE_NAME, CURRENT_VERSION, tokenMaker, userService)

	r.Run(":" + config.PORT)
}
