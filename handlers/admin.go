package handlers

import (
	"git.ssy.dk/noob/bingbong-go/models"
	"git.ssy.dk/noob/bingbong-go/templates"
	"git.ssy.dk/noob/bingbong-go/timing"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminDashboardHandler renders the admin dashboard
func AdminDashboardHandler(c *gin.Context) {
	t := c.MustGet("timing").(*timing.RenderTiming)
	db := c.MustGet("db").(*gorm.DB)

	// Get all users
	var users []models.User
	db.Find(&users)

	// Get all groups with related data
	var groups []models.UserGroup
	db.Preload("Creator").Preload("Members").Find(&groups)

	t.StartTemplate()
	templates.AdminDashboard(t, users, groups).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}
