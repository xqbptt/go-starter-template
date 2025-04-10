package apiUtils

import (
	"backend/api/middleware"
	"backend/token"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetContextFromGinContext(ginCtx *gin.Context) context.Context {
	// set authenticationPayloadKey and return context
	value, exists := ginCtx.Get(fmt.Sprint(middleware.AuthenticationPayloadKey))
	if !exists {
		value = &token.Payload{}
	}

	return context.WithValue(
		ginCtx.Request.Context(),
		middleware.AuthenticationPayloadKey,
		value.(*token.Payload))
}
