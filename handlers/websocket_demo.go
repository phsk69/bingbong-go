package handlers

import (
	"git.ssy.dk/noob/bingbong-go/templates"
	"github.com/gin-gonic/gin"
)

// WebSocketDemoHandler renders the websocket demo page
func WebSocketDemoHandler(c *gin.Context) {
	component := templates.WebSocketDemo()
	component.Render(c.Request.Context(), c.Writer)
}
