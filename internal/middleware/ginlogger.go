package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitistack/ipam-api/internal/logger"
)

func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		latency := time.Since(start)

		logger.HTTP.Infow("http request",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"ip", c.ClientIP(),
			"latency", latency.String(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

func ZapErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		for _, err := range c.Errors {
			logger.HTTP.Errorw("request error",
				"error", err.Err,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
			)
		}
	}
}
