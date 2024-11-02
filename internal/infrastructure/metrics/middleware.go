package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		// Record metrics
		ScheduleOpsTotal.WithLabelValues(c.Request.Method).Inc()
		ScheduleOpsDuration.WithLabelValues(c.Request.Method).Observe(duration)
	}
}
