package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewApiKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Api-Key") != apiKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "bad api key"})

			return
		}

		c.Next()
	}
}
