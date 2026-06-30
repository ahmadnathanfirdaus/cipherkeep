package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RequestID assigns a request id (honoring an inbound X-Request-ID header) and
// echoes it back in the response.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(ContextRequestID, rid)
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Next()
	}
}

// Recovery recovers from panics, logs the error, and returns a 500 envelope.
func Recovery(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(logrus.Fields{
					"request_id": RequestIDFromContext(c),
					"panic":      r,
					"path":       c.Request.URL.Path,
				}).Error("recovered from panic")
				c.AbortWithStatusJSON(500, gin.H{
					"error": gin.H{
						"code":    "INTERNAL",
						"message": "An internal error occurred",
					},
				})
			}
		}()
		c.Next()
	}
}

// BodyLimit caps the request body size to maxBytes. Reads beyond the limit fail,
// so handlers binding the body return a 400 instead of buffering unbounded input.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

// RequestLogger logs each request with structured fields. It never logs bodies.
func RequestLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		fields := logrus.Fields{
			"request_id": RequestIDFromContext(c),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"latency_ms": time.Since(start).Milliseconds(),
			"client_ip":  c.ClientIP(),
		}
		if user, ok := UserFromContext(c); ok {
			fields["user_id"] = user.ID
		}

		entry := log.WithFields(fields)
		status := c.Writer.Status()
		switch {
		case status >= 500:
			entry.Error("request completed")
		case status >= 400:
			entry.Warn("request completed")
		default:
			entry.Info("request completed")
		}
	}
}
