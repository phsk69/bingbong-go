// routes/router.go
package routes

import (
	"net/http"

	"git.ssy.dk/noob/snakey-go/v2/handlers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Router struct {
	db     *gorm.DB
	engine *gin.Engine
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

	return router
}

func (r *Router) SetupRoutes() {
	// Health check
	r.engine.GET("/ping", handlers.PingHandler)

	api := r.engine.Group("/api")
	v1 := api.Group("/v1")
	{
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
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
