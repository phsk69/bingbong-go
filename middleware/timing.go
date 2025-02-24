package middleware

import (
	"git.ssy.dk/noob/bingbong-go/timing"
	"github.com/gin-gonic/gin"
)

// TimingMiddleware adds timing information to HTML responses
func TimingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create timing object
		t := timing.NewRenderTiming()

		// Store in context for handlers
		c.Set("timing", t)

		// Wrap the response writer with our timing wrapper
		originalWriter := c.Writer
		timingWriter := timing.NewTimingResponseWriter(originalWriter, t)
		c.Writer = timingWriter

		// Process request
		c.Next()

		// Make sure we flush the content before finishing
		timingWriter.Flush()
	}
}
