package routes

import (
	"net/http"

	"git.ssy.dk/noob/bingbong-go/handlers"
	"git.ssy.dk/noob/bingbong-go/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Router struct {
	db     *gorm.DB
	engine *gin.Engine
	wsHub  *handlers.DistributedHub
}

func NewRouter(db *gorm.DB) *Router {
	router := &Router{
		db:     db,
		engine: gin.Default(),
	}

	// Add DB middleware
	router.engine.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Add timing middleware
	router.engine.Use(middleware.TimingMiddleware())

	return router
}

// SetHub sets the WebSocket hub for the router
func (r *Router) SetHub(hub *handlers.DistributedHub) {
	r.wsHub = hub
	// Add WebSocket hub middleware
	r.engine.Use(func(c *gin.Context) {
		c.Set("hub", r.wsHub)
		c.Next()
	})
}

func (r *Router) SetupRoutes() {
	// Authentication routes
	r.engine.GET("/login", handlers.LoginPageHandler)
	r.engine.GET("/logout", handlers.LogoutHandler)

	// WebSocket routes
	r.engine.GET("/ws", func(c *gin.Context) {
		if r.wsHub != nil {
			handlers.HandleWebSocket(c)
		} else {
			c.String(http.StatusServiceUnavailable, "WebSocket service not available")
		}
	})
	r.engine.GET("/demo", handlers.WebSocketDemoHandler)

	// Serve static files
	r.engine.Static("/static", "./static")

	// Homepage route
	r.engine.GET("/", handlers.HomeHandler)

	// Health check endpoints
	r.engine.GET("/ping", handlers.PingHandler)
	r.engine.GET("/healthz", func(c *gin.Context) {
		// Enhanced health check that includes Redis status
		if r.wsHub != nil && r.wsHub.IsHealthy() {
			handlers.HealthzHandler(c)
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"details": "WebSocket hub not available or unhealthy",
			})
		}
	})

	// API routes
	api := r.engine.Group("/api")
	v1 := api.Group("/v1")
	{
		// Auth API endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.LoginHandler)
		}

		// Public API endpoints
		users := v1.Group("/users")
		{
			users.POST("/", handlers.CreateUser)
			users.GET("/", handlers.GetUsers)
			users.GET("/:id", handlers.GetUser)
			users.PUT("/:id", handlers.UpdateUser)
			users.DELETE("/:id", handlers.DeleteUser)
		}

		groups := v1.Group("/groups")
		{
			groups.POST("/", handlers.CreateGroup)
			groups.GET("/", handlers.GetGroups)
			groups.GET("/:id", handlers.GetGroup)
			groups.PUT("/:id", handlers.UpdateGroup)
			groups.DELETE("/:id", handlers.DeleteGroup)
		}

		// Protected admin API endpoints
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminAuthMiddleware())
		{
			// Admin user management
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("/", handlers.AdminGetUsersHandler)
				adminUsers.GET("/new", handlers.AdminGetUserFormHandler)
				adminUsers.GET("/:id/edit", handlers.AdminGetUserEditFormHandler)
				adminUsers.POST("/", handlers.AdminCreateUserHandler)
				adminUsers.PUT("/:id", handlers.AdminUpdateUserHandler)
				adminUsers.DELETE("/:id", handlers.AdminDeleteUserHandler)
			}

			// Admin group management could be added similarly here
		}

		// WebSocket related APIs
		websocket := v1.Group("/websocket")
		{
			websocket.GET("/stats", func(c *gin.Context) {
				if r.wsHub != nil {
					c.JSON(http.StatusOK, gin.H{
						"connections": r.wsHub.GetStats(),
					})
				} else {
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error": "WebSocket service not available",
					})
				}
			})
		}
	}

	// Admin routes - protected and rendered pages
	adminRoutes := r.engine.Group("/admin")
	adminRoutes.Use(middleware.AuthMiddleware(), middleware.AdminAuthMiddleware())
	{
		adminRoutes.GET("/", handlers.AdminDashboardHandler) // Renders the admin dashboard
		// Add more admin UI routes here as needed
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
