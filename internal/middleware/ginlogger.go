package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitistack/ipam-api/internal/logger"
)

func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		// query := c.Request.URL.RawQuery

		var requestBody string
		if c.Request.Body != nil {
			// Read the body and mask the secret
			bodyBytes, err := c.GetRawData()
			if err == nil {
				masked := maskSecretField(bodyBytes)
				requestBody = string(masked)
				// Restore the io.ReadCloser to its original state
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		c.Next()
		responseTime := time.Since(start)
		responseTimeMs := float64(responseTime.Microseconds()) / 1000.0
		logger.HTTP.Infow("http request",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			// "query", query,
			"ip", c.ClientIP(),
			"response_time", responseTime.String(),
			"response_time_ms", responseTimeMs,
			"user_agent", c.Request.UserAgent(),
			"body", requestBody,
		)
	}
}

func ZapErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := c.GetRawData()
			if err == nil {
				masked := maskSecretField(bodyBytes)
				requestBody = string(masked)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		c.Next()
		for _, err := range c.Errors {
			logger.HTTP.Errorw("request error",
				"error", err.Err,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
				"method", c.Request.Method,
				"ip", c.ClientIP(),
				"body", requestBody,
			)
		}
	}
}

// maskSecretField masks the value of the "secret" field in a JSON body.
func maskSecretField(body []byte) []byte {
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return body
	}
	if val, ok := m["secret"]; ok && val != nil {
		m["secret"] = "*********"
	}
	masked, err := json.Marshal(m)
	if err != nil {
		return body
	}
	return masked
}
