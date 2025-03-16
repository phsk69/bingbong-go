package handlers

import (
	"net/http"
	"strconv"

	"git.ssy.dk/noob/bingbong-go/models"
	"git.ssy.dk/noob/bingbong-go/templates"
	"git.ssy.dk/noob/bingbong-go/timing"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminGetGroupsHandler handles getting the groups list for the admin panel
func AdminGetGroupsHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	var groups []models.UserGroup

	// Preload relationships for proper display
	if err := db.Preload("Creator").Preload("Members").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	t.StartTemplate()
	templates.AdminGroupsList(groups).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// AdminGetGroupFormHandler returns the form for creating a new group
func AdminGetGroupFormHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)

	// Get all users for the member selection dropdown
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	t.StartTemplate()
	templates.AdminGroupForm(models.UserGroup{}, users, nil).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// AdminGetGroupEditFormHandler returns the form for editing a group
func AdminGetGroupEditFormHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var group models.UserGroup
	if err := db.Preload("Members.User").First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Get all users for the member selection dropdown
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Create a map of current member IDs for easier checking in the template
	memberIDs := make(map[uint]bool)
	for _, member := range group.Members {
		memberIDs[member.UserID] = true
	}

	t.StartTemplate()
	templates.AdminGroupForm(group, users, memberIDs).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// AdminCreateGroupHandler creates a new group
func AdminCreateGroupHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	var groupRequest struct {
		Name        string   `form:"name" binding:"required"`
		Description string   `form:"description"`
		MemberIDs   []string `form:"member_ids[]"`
	}

	if err := c.ShouldBind(&groupRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if group name already exists
	var existingGroup models.UserGroup
	if db.Where("name = ?", groupRequest.Name).First(&existingGroup).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group name already exists"})
		return
	}

	// Create the group
	group := models.UserGroup{
		Name:        groupRequest.Name,
		Description: groupRequest.Description,
		CreatedByID: userID,
	}

	if err := db.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	// Add members to the group if any are selected
	if len(groupRequest.MemberIDs) > 0 {
		// Convert string IDs to uint
		var memberIDs []uint
		for _, idStr := range groupRequest.MemberIDs {
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				continue // Skip invalid IDs
			}
			memberIDs = append(memberIDs, uint(id))
		}

		if len(memberIDs) > 0 {
			// Create UserGroupMember entries for each selected user
			for _, memberID := range memberIDs {
				membership := models.UserGroupMember{
					UserID:  memberID,
					GroupID: group.ID,
				}
				if err := db.Create(&membership).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to group"})
					return
				}
			}
		}
	}

	// Return the updated group list
	var groups []models.UserGroup
	if err := db.Preload("Creator").Preload("Members.User").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	t := c.MustGet("timing").(*timing.RenderTiming)
	t.StartTemplate()
	templates.AdminGroupsList(groups).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// AdminUpdateGroupHandler updates an existing group
func AdminUpdateGroupHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// Get form values
	name := c.PostForm("name")
	description := c.PostForm("description")
	memberIDs := c.PostFormArray("member_ids[]")

	// Validation
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group name is required"})
		return
	}

	// Get the existing group
	var group models.UserGroup
	if err := db.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Update group fields
	group.Name = name
	group.Description = description

	// Update the group
	if err := db.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	// Convert member IDs from string to uint
	var numericMemberIDs []uint
	for _, idStr := range memberIDs {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			continue // Skip invalid IDs
		}
		numericMemberIDs = append(numericMemberIDs, uint(id))
	}

	// Update group members - first clear existing members
	if err := db.Model(&group).Association("Members").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group members"})
		return
	}

	// Then add the new members
	if len(numericMemberIDs) > 0 {
		var members []models.User
		if err := db.Where("id IN ?", numericMemberIDs).Find(&members).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
			return
		}

		if err := db.Model(&group).Association("Members").Append(&members); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add members to group"})
			return
		}
	}

	// Return the updated group list
	var groups []models.UserGroup
	db.Preload("Creator").Preload("Members").Find(&groups)

	t := c.MustGet("timing").(*timing.RenderTiming)
	t.StartTemplate()
	templates.AdminGroupsList(groups).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// AdminDeleteGroupHandler deletes a group
func AdminDeleteGroupHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// Check if group exists
	var group models.UserGroup
	if err := db.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Delete the group (will cascade delete related records)
	if err := db.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	// Get the updated group list and return it for UI update
	var groups []models.UserGroup
	db.Preload("Creator").Preload("Members").Find(&groups)

	t.StartTemplate()
	templates.AdminGroupsList(groups).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}
