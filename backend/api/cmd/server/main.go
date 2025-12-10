// Package main is the entry point for the CloudCop API server.
package main

import (
	"log"

	"cloudcop/api/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/health", handlers.Health)
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
