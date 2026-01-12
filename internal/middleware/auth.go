package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// TokenAuth validates the request against a predefined token from auth_token.secret
func TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedToken := viper.GetString("auth.token") // strings.TrimSpace(string(tokenBytes))

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		// Support both "Bearer <token>" and plain token
		token := authHeader
		if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
			token = after
		}

		// Validate token
		if token != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
