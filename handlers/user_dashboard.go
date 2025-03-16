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

// UserDashboardHandler renders the main user dashboard
func UserDashboardHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	t.StartTemplate()
	templates.UserDashboard(user, "account").Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// UserGroupsHandler renders the groups tab of the user dashboard
func UserGroupsHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	t.StartTemplate()
	templates.UserDashboard(user, "groups").Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// GetUserGroupsDataHandler fetches and renders just the groups list - updated
func GetUserGroupsDataHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Get groups where user is a member or creator
	var groups []models.UserGroup
	if err := db.Preload("Creator").Preload("Members.User").
		Joins("LEFT JOIN user_group_members ON user_groups.id = user_group_members.group_id").
		Where("user_groups.created_by_id = ? OR user_group_members.user_id = ?", userID, userID).
		Group("user_groups.id").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	t.StartTemplate()
	templates.UserGroups(user, groups).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// DeleteGroupHandler deletes a user group
func DeleteGroupHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// Verify the group exists and the user is the creator
	var group models.UserGroup
	if err := db.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Only the creator can delete the group
	if group.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this group"})
		return
	}

	// Delete the group
	if err := db.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	// Get user data
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Return the updated groups list
	var groups []models.UserGroup
	if err := db.Preload("Creator").Preload("Members.User").
		Joins("LEFT JOIN user_group_members ON user_groups.id = user_group_members.group_id").
		Where("user_groups.created_by_id = ? OR user_group_members.user_id = ?", userID, userID).
		Group("user_groups.id").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	t := c.MustGet("timing").(*timing.RenderTiming)
	t.StartTemplate()
	templates.UserGroups(user, groups).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// UserInvitesHandler renders the invitations tab of the user dashboard
func UserInvitesHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	t.StartTemplate()
	templates.UserDashboard(user, "invites").Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// GetUserInvitesDataHandler fetches and renders just the invitations lists
func GetUserInvitesDataHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Get sent invites
	var sentInvites []models.UserGroupInvite
	if err := db.Preload("Group").Preload("Invitee").
		Where("invite_initiator_id = ?", userID).
		Find(&sentInvites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sent invites"})
		return
	}

	// Get received invites
	var receivedInvites []models.UserGroupInvite
	if err := db.Preload("Group").Preload("Initiator").
		Where("invitee_id = ? AND accepted = ?", userID, false).
		Find(&receivedInvites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch received invites"})
		return
	}

	t.StartTemplate()
	templates.UserInvites(user, sentInvites, receivedInvites).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// UpdateUserPasswordHandler updates the user's password
func UpdateUserPasswordHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	var passwordRequest struct {
		CurrentPassword string `form:"current_password" binding:"required"`
		NewPassword     string `form:"new_password" binding:"required"`
		ConfirmPassword string `form:"confirm_password" binding:"required"`
	}

	if err := c.ShouldBind(&passwordRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify passwords match
	if passwordRequest.NewPassword != passwordRequest.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New passwords do not match"})
		return
	}

	// Get the user
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordRequest.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordRequest.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update the password
	user.Password = string(hashedPassword)
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Return success message
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// UpdateUserPublicKeyHandler updates the user's public key
func UpdateUserPublicKeyHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	var keyRequest struct {
		PublicKey string `form:"public_key"`
	}

	if err := c.ShouldBind(&keyRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the user
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Update the public key
	user.PublicKey = keyRequest.PublicKey
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update public key"})
		return
	}

	// Return success message
	c.JSON(http.StatusOK, gin.H{"message": "Public key updated successfully"})
}

// GetCreateGroupFormHandler returns the form for creating a new group
func GetCreateGroupFormHandler(c *gin.Context) {
	t := c.MustGet("timing").(*timing.RenderTiming)

	t.StartTemplate()
	templates.CreateGroupForm().Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// GetEditGroupFormHandler returns the form for editing a group
func GetEditGroupFormHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	userID := c.MustGet("userID").(uint)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var group models.UserGroup
	if err := db.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Verify user owns the group
	if group.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to edit this group"})
		return
	}

	t.StartTemplate()
	templates.EditGroupForm(group).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// GetGroupDetailHandler returns the detailed view of a group
func GetGroupDetailHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	userID := c.MustGet("userID").(uint)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var group models.UserGroup
	if err := db.Preload("Creator").Preload("Members.User").First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Verify user is a member of the group or the creator
	isMember := false
	if group.CreatedByID == userID {
		isMember = true
	} else {
		for _, member := range group.Members {
			if member.UserID == userID {
				isMember = true
				break
			}
		}
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this group"})
		return
	}

	t.StartTemplate()
	templates.GroupDetail(group, userID).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// CreateGroupHandler creates a new group - updated
func CreateGroupHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	var groupRequest struct {
		Name        string `form:"name" binding:"required"`
		Description string `form:"description"`
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

	// Add creator as a member
	membership := models.UserGroupMember{
		UserID:  userID,
		GroupID: group.ID,
	}
	if err := db.Create(&membership).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add you as a member"})
		return
	}

	// Get user data
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Return the updated groups list
	var groups []models.UserGroup
	if err := db.Preload("Creator").Preload("Members.User").
		Joins("LEFT JOIN user_group_members ON user_groups.id = user_group_members.group_id").
		Where("user_groups.created_by_id = ? OR user_group_members.user_id = ?", userID, userID).
		Group("user_groups.id").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	t := c.MustGet("timing").(*timing.RenderTiming)
	t.StartTemplate()
	templates.UserGroups(user, groups).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// UpdateGroupHandler updates an existing group - updated
func UpdateGroupHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var groupRequest struct {
		Name        string `form:"name" binding:"required"`
		Description string `form:"description"`
	}

	if err := c.ShouldBind(&groupRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the existing group
	var group models.UserGroup
	if err := db.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Verify user owns the group
	if group.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to edit this group"})
		return
	}

	// Update group fields
	group.Name = groupRequest.Name
	group.Description = groupRequest.Description

	// Update the group
	if err := db.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	// Get user data for the response
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Return the updated groups list
	var groups []models.UserGroup
	if err := db.Preload("Creator").Preload("Members.User").
		Joins("LEFT JOIN user_group_members ON user_groups.id = user_group_members.group_id").
		Where("user_groups.created_by_id = ? OR user_group_members.user_id = ?", userID, userID).
		Group("user_groups.id").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	t := c.MustGet("timing").(*timing.RenderTiming)
	t.StartTemplate()
	templates.UserGroups(user, groups).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// GetInviteUserFormHandler returns the form for inviting a user to a group
func GetInviteUserFormHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	userID := c.MustGet("userID").(uint)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// Verify user owns the group
	var group models.UserGroup
	if err := db.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	if group.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to invite users to this group"})
		return
	}

	// Get users who are not already members of the group
	var users []models.User
	if err := db.Where("id NOT IN (SELECT user_id FROM user_group_members WHERE group_id = ?)", groupID).
		Where("id NOT IN (SELECT invitee_id FROM user_group_invites WHERE group_id = ? AND accepted = ?)", groupID, false).
		Where("id != ?", userID). // Exclude current user
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	t.StartTemplate()
	templates.InviteUserForm(uint(groupID), users).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// InviteUserToGroupHandler invites a user to a group - Updated with notifications
func InviteUserToGroupHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	hubInterface, hubExists := c.Get("hub")
	userID := c.MustGet("userID").(uint)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var inviteRequest struct {
		InviteeID string `form:"invitee_id" binding:"required"`
	}

	if err := c.ShouldBind(&inviteRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inviteeID, err := strconv.ParseUint(inviteRequest.InviteeID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitee ID"})
		return
	}

	// Get the current user
	var currentUser models.User
	if err := db.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch current user"})
		return
	}

	// Verify user owns the group
	var group models.UserGroup
	if err := db.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	if group.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to invite users to this group"})
		return
	}

	// Check if user is already a member
	var existingMembership models.UserGroupMember
	if err := db.Where("user_id = ? AND group_id = ?", inviteeID, groupID).First(&existingMembership).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already a member of this group"})
		return
	}

	// Check if invitation already exists
	var existingInvite models.UserGroupInvite
	if err := db.Where("invitee_id = ? AND group_id = ? AND accepted = ?", inviteeID, groupID, false).
		First(&existingInvite).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User has already been invited to this group"})
		return
	}

	// Create the invitation
	invite := models.UserGroupInvite{
		GroupID:           uint(groupID),
		InviteInitiatorID: userID,
		InviteeID:         uint(inviteeID),
		Accepted:          false,
	}

	if err := db.Create(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
		return
	}

	// Send a WebSocket notification if hub exists
	if hubExists {
		if hub, ok := hubInterface.(*DistributedHub); ok {
			hub.SendGroupInviteNotification(
				uint(inviteeID),
				currentUser.Username,
				group.Name,
				invite.ID,
			)
		}
	}

	// Return the updated sent invites list
	var sentInvites []models.UserGroupInvite
	if err := db.Preload("Group").Preload("Invitee").
		Where("invite_initiator_id = ?", userID).
		Find(&sentInvites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sent invites"})
		return
	}

	t := c.MustGet("timing").(*timing.RenderTiming)
	t.StartTemplate()
	templates.UserInvites(models.User{ID: userID}, sentInvites, nil).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// AcceptInviteHandler accepts a group invitation
func AcceptInviteHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	inviteID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invite ID"})
		return
	}

	// Get the invitation
	var invite models.UserGroupInvite
	if err := db.Preload("Group").First(&invite, inviteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
		return
	}

	// Verify user is the invitee
	if invite.InviteeID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to accept this invitation"})
		return
	}

	// Mark the invitation as accepted
	invite.Accepted = true
	if err := db.Save(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invitation"})
		return
	}

	// Add user to the group
	membership := models.UserGroupMember{
		UserID:  userID,
		GroupID: invite.GroupID,
	}
	if err := db.Create(&membership).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add you to the group"})
		return
	}

	// Return the updated received invites list
	var receivedInvites []models.UserGroupInvite
	if err := db.Preload("Group").Preload("Initiator").
		Where("invitee_id = ? AND accepted = ?", userID, false).
		Find(&receivedInvites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch received invites"})
		return
	}

	t := c.MustGet("timing").(*timing.RenderTiming)
	t.StartTemplate()
	templates.UserInvites(models.User{ID: userID}, nil, receivedInvites).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// DeclineInviteHandler declines a group invitation
func DeclineInviteHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	inviteID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invite ID"})
		return
	}

	// Get the invitation
	var invite models.UserGroupInvite
	if err := db.First(&invite, inviteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
		return
	}

	// Verify user is the invitee or the initiator
	if invite.InviteeID != userID && invite.InviteInitiatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to decline this invitation"})
		return
	}

	// Delete the invitation
	if err := db.Delete(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decline invitation"})
		return
	}

	// Determine which list to refresh based on who declined
	if invite.InviteeID == userID {
		// User is declining an invitation they received
		var receivedInvites []models.UserGroupInvite
		if err := db.Preload("Group").Preload("Initiator").
			Where("invitee_id = ? AND accepted = ?", userID, false).
			Find(&receivedInvites).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch received invites"})
			return
		}

		t := c.MustGet("timing").(*timing.RenderTiming)
		t.StartTemplate()
		templates.UserInvites(models.User{ID: userID}, nil, receivedInvites).Render(c.Request.Context(), c.Writer)
		t.EndTemplate()
	} else {
		// User is canceling an invitation they sent
		var sentInvites []models.UserGroupInvite
		if err := db.Preload("Group").Preload("Invitee").
			Where("invite_initiator_id = ?", userID).
			Find(&sentInvites).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sent invites"})
			return
		}

		t := c.MustGet("timing").(*timing.RenderTiming)
		t.StartTemplate()
		templates.UserInvites(models.User{ID: userID}, sentInvites, nil).Render(c.Request.Context(), c.Writer)
		t.EndTemplate()
	}
}

// GetUserAccountHandler fetches and renders just the account settings
func GetUserAccountHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	t.StartTemplate()
	templates.UserAccountSettings(user).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}

// RemoveGroupMemberHandler removes a member from a group
func RemoveGroupMemberHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	t := c.MustGet("timing").(*timing.RenderTiming)
	userID := c.MustGet("userID").(uint)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	memberID, err := strconv.ParseUint(c.Param("member_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	// Verify the group exists and the current user is the creator
	var group models.UserGroup
	if err := db.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Only the creator can remove members
	if group.CreatedByID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to manage this group's members"})
		return
	}

	// Cannot remove the creator (owner)
	if uint(memberID) == group.CreatedByID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove the group owner"})
		return
	}

	// Delete the membership
	result := db.Where("user_id = ? AND group_id = ?", memberID, groupID).Delete(&models.UserGroupMember{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found in this group"})
		return
	}

	// Reload the group with members
	if err := db.Preload("Creator").Preload("Members.User").First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload group data"})
		return
	}

	// Render the updated group detail
	t.StartTemplate()
	templates.GroupDetail(group, userID).Render(c.Request.Context(), c.Writer)
	t.EndTemplate()
}
