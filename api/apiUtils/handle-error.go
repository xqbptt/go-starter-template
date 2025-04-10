package apiUtils

import (
	"backend/dto"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SendErrorResponse(c *gin.Context, err error) {
	var httpErr *dto.Error
	if errors.As(err, &httpErr) {
		c.JSON(httpErr.Code, err)
		return
	}
	c.JSON(http.StatusInternalServerError, dto.NewError(err.Error()))
}
