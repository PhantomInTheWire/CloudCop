// Package handlers contains HTTP request handlers for the API.
package handlers

import "github.com/gin-gonic/gin"

// Health handles health check requests.
func Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "api",
	})
}
