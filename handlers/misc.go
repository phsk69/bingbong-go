package handlers

import (
	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func HealthzHandler(c *gin.Context) {
	// Check if the hub is healthy
	if c.MustGet("hub").(*DistributedHub).IsHealthy() {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	} else {
		c.JSON(500, gin.H{
			"status": "error",
		})
	}
}
