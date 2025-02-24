package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"git.ssy.dk/noob/bingbong-go/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Key constants for JWT
const (
	TokenTTL = 24 * time.Hour
)

var secretKey string

func InitSecretKey() {
	secretKey = os.Getenv("JWT_SECRET")
	if secretKey == "" {
		panic("JWT_SECRET environment variable is not set")
	}
}

// Claims struct for JWT
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(user *models.User, db *gorm.DB) (string, error) {
	// Check if user is an admin
	var adminAccess models.AdminGroupMember
	isAdmin := db.Where("user_id = ? AND active = ?", user.ID, true).First(&adminAccess).Error == nil

	// Create claims with user information
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// AuthMiddleware checks if the user is authenticated
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get auth token from header
		authHeader := c.GetHeader("Authorization")

		// Check for JWT in cookie as fallback
		tokenCookie, err := c.Cookie("auth_token")

		// If no auth header and no valid cookie, redirect to login
		if (authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ")) && (err != nil || tokenCookie == "") {
			// Store the original URL for redirection after login
			originalURL := c.Request.URL.String()

			// Encode the original URL to use as a query parameter
			redirectParam := url.QueryEscape(originalURL)

			// Redirect to login page with the return URL
			c.Redirect(http.StatusFound, "/login?redirect="+redirectParam)
			c.Abort()
			return
		}

		// Extract token from header or cookie
		var token string
		if authHeader != "" {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			token = tokenCookie
		}

		// Parse and validate the token
		claims := &Claims{}
		jwtToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil || !jwtToken.Valid {
			// Store the original URL for redirection after login
			originalURL := c.Request.URL.String()

			// Encode the original URL to use as a query parameter
			redirectParam := url.QueryEscape(originalURL)

			// Redirect to login page with the return URL
			c.Redirect(http.StatusFound, "/login?redirect="+redirectParam)
			c.Abort()
			return
		}

		// Set user info in context for downstream handlers
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("isAdmin", claims.IsAdmin)

		c.Next()
	}
}

// AdminAuthMiddleware ensures the user is an admin
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get isAdmin from context (set by AuthMiddleware)
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
