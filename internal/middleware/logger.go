package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logger() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		requestID := uuid.New().String()
		ctx.Set("request_id", requestID)
		ctx.Header("X-Request-ID", requestID)

		start := time.Now()
		ctx.Next() // next jhandler
		duration := time.Since(start)

		status := ctx.Writer.Status()

		logLine := map[string]interface{}{
			"request_id": requestID,
			"method":     ctx.Request.Method,
			"path":       ctx.Request.URL.Path,
			"status":     status,
			"duration_ms": duration.Milliseconds(),
			"ip":         ctx.ClientIP(),	
		}
		if status >= 500 {
			log.Printf("[ERROR] %+v", logLine)
		} else if status == 401 || status == 403 {
			log.Printf("[AUTH_FAIL] %+v", logLine)
		} else {
			log.Printf("[INFO] %+v", logLine)
		}
	}
}