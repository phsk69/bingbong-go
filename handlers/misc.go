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
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
