package handlers

import (
	"net/http"
	"strconv"

	"git.ssy.dk/noob/bingbong-go/models"
	"git.ssy.dk/noob/bingbong-go/templates"
	"git.ssy.dk/noob/bingbong-go/timing"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminGetUsersHandler handles getting the user list for the admin panel
func AdminGetUsersHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	t.StartTemplate()
	templates.AdminUsersList(users).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// AdminGetUserFormHandler returns the form for creating a new user
func AdminGetUserFormHandler(c *gin.Context) {
	t := c.MustGet("timing").(*timing.RenderTiming)

	t.StartTemplate()
	templates.AdminUserForm(models.User{}, false).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// AdminGetUserEditFormHandler returns the form for editing a user
func AdminGetUserEditFormHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is admin
	isAdmin := db.Where("user_id = ? AND active = ?", user.ID, true).First(&models.AdminGroupMember{}).Error == nil

	t.StartTemplate()
	templates.AdminUserForm(user, isAdmin).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// AdminCreateUserHandler creates a new user
func AdminCreateUserHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var userRequest struct {
		Username string `form:"username" binding:"required"`
		Email    string `form:"email" binding:"required,email"`
		Password string `form:"password" binding:"required,min=6"`
		IsAdmin  bool   `form:"is_admin"`
	}

	if err := c.ShouldBind(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username or email already exists
	var existingUser models.User
	if db.Where("username = ? OR email = ?", userRequest.Username, userRequest.Email).First(&existingUser).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create the user
	user := models.User{
		Username: userRequest.Username,
		Email:    userRequest.Email,
		Password: string(hashedPassword),
		Active:   true,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// If user should be an admin, add them to the admin group
	if userRequest.IsAdmin {
		adminAccess := models.AdminGroupMember{
			UserID: user.ID,
			Active: true,
		}

		if err := db.Create(&adminAccess).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set admin status"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"isAdmin":  userRequest.IsAdmin,
	})
}

// AdminUpdateUserHandler updates an existing user
func AdminUpdateUserHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get form values
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	active := c.PostForm("active") == "on"
	isAdmin := c.PostForm("is_admin") == "on"

	// Validation
	if username == "" || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and email are required"})
		return
	}

	// Get the existing user
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user fields
	user.Username = username
	user.Email = email
	user.Active = active

	// Update password if provided
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	// Update the user
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Handle admin status
	var adminAccess models.AdminGroupMember
	adminExists := db.Where("user_id = ?", user.ID).First(&adminAccess).Error == nil

	if isAdmin {
		// User should be an admin
		if !adminExists {
			// Create new admin access
			adminAccess = models.AdminGroupMember{
				UserID: user.ID,
				Active: true,
			}
			if err := db.Create(&adminAccess).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set admin status"})
				return
			}
		} else if !adminAccess.Active {
			// Update existing inactive admin access
			adminAccess.Active = true
			if err := db.Save(&adminAccess).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin status"})
				return
			}
		}
	} else if adminExists && adminAccess.Active {
		// User should not be an admin, but currently is
		adminAccess.Active = false
		if err := db.Save(&adminAccess).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin status"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"active":   user.Active,
		"isAdmin":  isAdmin,
	})
}

// AdminDeleteUserHandler deletes a user
func AdminDeleteUserHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if user exists
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete the user (will cascade delete related records)
	if err := db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Get the updated user list and return it for UI update
	var users []models.User
	db.Find(&users)

	t.StartTemplate()
	templates.AdminUsersList(users).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}
