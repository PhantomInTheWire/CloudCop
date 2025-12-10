package main

import (
	"cloudcop/api/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/health", handlers.Health)
	r.Run(":8080")
}
