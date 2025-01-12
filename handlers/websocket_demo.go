package handlers

import (
	"git.ssy.dk/noob/snakey-go/templates"
	"github.com/gin-gonic/gin"
)

func WebSocketDemoHandler(c *gin.Context) {
	component := templates.WebSocketDemo()
	component.Render(c.Request.Context(), c.Writer)
}
