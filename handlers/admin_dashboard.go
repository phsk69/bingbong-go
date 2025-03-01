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
	// Get timing object from context
	t := c.MustGet("timing").(*timing.RenderTiming)
	db := c.MustGet("db").(*gorm.DB)

	// Explicitly set content type for HTML
	c.Header("Content-Type", "text/html; charset=utf-8")

	// Get all users with their admin access records
	var users []models.User
	db.Preload("AdminAccess").Find(&users)

	// Get all groups with related data
	var groups []models.UserGroup
	db.Preload("Creator").Preload("Members").Find(&groups)

	// Start template timing
	t.StartTemplate()

	// Render the admin dashboard
	templates.AdminDashboard(t, users, groups).Render(c.Request.Context(), c.Writer)

	// End template timing
	t.EndTemplate()
}
