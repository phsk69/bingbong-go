package handlers

import (
	"net/http"
	"time"

	"git.ssy.dk/noob/bingbong-go/middleware"
	"git.ssy.dk/noob/bingbong-go/models"
	"git.ssy.dk/noob/bingbong-go/templates"
	"git.ssy.dk/noob/bingbong-go/timing"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginPageHandler renders the login page
func LoginPageHandler(c *gin.Context) {
	t, exists := c.Get("timing")
	if !exists {
		t = timing.NewRenderTiming() // Fallback if timing middleware isn't available
	}
	renderTiming := t.(*timing.RenderTiming)

	// Get redirect URL from query parameter
	redirect := c.Query("redirect")

	renderTiming.StartTemplate()
	templates.Login(renderTiming, redirect).Render(c.Request.Context(), c.Writer)
	renderTiming.EndTemplate()
}

// LoginHandler handles user authentication and token generation
func LoginHandler(c *gin.Context) {
	var loginRequest struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}
	if err := c.ShouldBind(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get the redirect URL from query parameter
	redirect := c.Query("redirect")

	// Get the database instance
	db := c.MustGet("db").(*gorm.DB)

	// Find the user
	var user models.User
	if err := db.Where("username = ? AND active = ?", loginRequest.Username, true).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare the password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login time
	db.Model(&user).Update("last_login", time.Now())

	// Generate JWT token
	token, err := middleware.GenerateToken(&user, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Check if it's an API request or a form submission
	if c.GetHeader("HX-Request") != "" {
		// HTMX request - set cookie and return JSON
		c.SetCookie("auth_token", token, int(middleware.TokenTTL.Seconds()), "/", "", false, true)

		response := gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"isAdmin":  db.Where("user_id = ? AND active = ?", user.ID, true).First(&models.AdminGroupMember{}).Error == nil,
			},
		}

		// Include redirect URL in the response if provided
		if redirect != "" {
			response["redirect"] = redirect
		}

		c.JSON(http.StatusOK, response)
	} else {
		// Regular form submission - set cookie and redirect
		c.SetCookie("auth_token", token, int(middleware.TokenTTL.Seconds()), "/", "", false, true)

		// Redirect to the provided URL or default to home
		if redirect != "" {
			c.Redirect(http.StatusFound, redirect)
		} else {
			c.Redirect(http.StatusFound, "/")
		}
	}
}

// LogoutHandler handles user logout
func LogoutHandler(c *gin.Context) {
	// Clear the auth cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)

	// Redirect to login page
	c.Redirect(http.StatusFound, "/login")
}
