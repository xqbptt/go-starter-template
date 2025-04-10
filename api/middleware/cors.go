package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(origins []string) gin.HandlerFunc {
	originStr := "*"

	return func(c *gin.Context) {
		if len(origins) > 0 {
			originStr = origins[0]
			originReq := c.Request.Header.Get("Origin")
			for _, r := range origins {
				if r == originReq {
					originStr = r
				}
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", originStr)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, ngrok-skip-browser-warning")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
