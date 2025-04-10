package utils

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseToInt64OrNotFound(ctx *gin.Context, paramName string) (int64, error) {
	someID, err := strconv.ParseInt(ctx.Param(paramName), 10, 64)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return 0, errors.New("given path parameter is not integer")
	}
	return someID, nil
}
