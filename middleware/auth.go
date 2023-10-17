package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		a := c.Request.Header.Get("X-API-Key")
		if a != apiKey {
			log.Println("Invalid X-API-Key received!")
			return
		}

		c.Next()
	}
}
