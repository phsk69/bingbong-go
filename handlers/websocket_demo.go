package handlers

import (
	"net/http"

	"git.ssy.dk/noob/bingbong-go/templates"
	"git.ssy.dk/noob/bingbong-go/timing"
	"github.com/gin-gonic/gin"
)

// WebSocketDemoHandler renders the WebSocket demo page
func WebSocketDemoHandler(c *gin.Context) {
	// Get timing from context
	timingValue, exists := c.Get("timing")
	var t *timing.RenderTiming
	if exists {
		t = timingValue.(*timing.RenderTiming)
	} else {
		// Create a new timing object if none exists
		t = timing.NewRenderTiming()
	}

	// Create the template component
	component := templates.WebSocketDemo(t)

	// Set proper content type
	c.Header("Content-Type", "text/html; charset=utf-8")

	// Start template timing
	t.StartTemplate()

	// Render the template
	err := component.Render(c.Request.Context(), c.Writer)

	// End template timing AFTER rendering
	t.EndTemplate()

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
