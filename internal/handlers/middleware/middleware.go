package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Обработка запроса
		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		logger.Info("HTTP request",
			zap.String("path", path),
			zap.String("method", method),
			zap.Int("status", status),
			zap.Duration("duration", duration),
		)
	}
}

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Проверяем наличие ошибок
		if len(c.Errors) > 0 {
			// Берем последнюю ошибку
			err := c.Errors.Last()

			// Отправляем соответствующий ответ
			c.JSON(c.Writer.Status(), gin.H{
				"error": err.Error(),
			})
		}
	}
}
