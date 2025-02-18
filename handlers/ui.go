package handlers

import (
	"git.ssy.dk/noob/bingbong-go/templates"
	"github.com/gin-gonic/gin"
)

func HomeHandler(c *gin.Context) {
	// Create a new component
	component := templates.HomePage()

	// Render the component
	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.String(500, "Internal Server Error")
		return
	}
}
