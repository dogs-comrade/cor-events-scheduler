// internal/handlers/middleware/metrics.go
package middleware

import (
	"fmt"
	"time"

	"cor-events-scheduler/internal/metrics"

	"github.com/gin-gonic/gin"
)

type metricsMiddleware struct {
	skipPaths map[string]bool
}

func NewMetricsMiddleware(skipPaths ...string) gin.HandlerFunc {
	mw := &metricsMiddleware{
		skipPaths: make(map[string]bool),
	}

	for _, path := range skipPaths {
		mw.skipPaths[path] = true
	}

	return mw.Handle
}

func (mw *metricsMiddleware) Handle(c *gin.Context) {
	if mw.skipPaths[c.Request.URL.Path] {
		c.Next()
		return
	}

	start := time.Now()
	path := c.Request.URL.Path
	method := c.Request.Method

	metrics.HttpRequestsInFlight.Inc()

	if c.Request.ContentLength > 0 {
		metrics.HttpRequestSize.WithLabelValues(method, path).Observe(float64(c.Request.ContentLength))
	}

	responseWriter := &responseWriterMetrics{ResponseWriter: c.Writer}
	c.Writer = responseWriter

	c.Next()

	metrics.HttpRequestsInFlight.Dec()

	duration := time.Since(start)
	status := responseWriter.Status()

	metrics.HttpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
	metrics.HttpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())

	if responseWriter.written > 0 {
		metrics.HttpResponseSize.WithLabelValues(
			method,
			path,
			fmt.Sprintf("%d", status),
		).Observe(float64(responseWriter.written))
	}
}

type responseWriterMetrics struct {
	gin.ResponseWriter
	written int64
}

func (w *responseWriterMetrics) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.written += int64(n)
	return n, err
}
