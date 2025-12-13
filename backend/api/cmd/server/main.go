// Package main is the entry point for the CloudCop API server.
package main

import (
	"log"

	"cloudcop/api/internal/awsauth"
	"cloudcop/api/internal/handlers"

	"github.com/gin-gonic/gin"
)

// main initializes AWS authentication and a credential cache, registers HTTP handlers for health checks and account management under /api/accounts, and starts the Gin HTTP server on :8080.
func main() {
	/*
		Initialize the AWS authentication service which handles STS AssumeRole operations.
		This service manages connections to AWS accounts using temporary credentials
		with ExternalID validation for security.
	*/
	auth, err := awsauth.NewAWSAuth()
	if err != nil {
		log.Fatalf("Failed to initialize AWS auth: %v", err)
	}

	/*
		Create a credential cache to store and automatically refresh temporary AWS credentials.
		The cache prevents unnecessary STS API calls and ensures credentials are refreshed
		before they expire (5 minutes before expiration by default).
	*/
	cache := awsauth.NewCredentialCache(auth)

	/*
		Initialize the accounts handler which provides HTTP endpoints for managing
		AWS account connections, including verify, connect, list, and disconnect operations.
	*/
	accountsHandler := handlers.NewAccountsHandler(auth, cache)

	/*
		Setup the Gin web framework router with health check and account management endpoints.
		All account endpoints are grouped under /api/accounts for RESTful consistency.
	*/
	r := gin.Default()

	r.GET("/health", handlers.Health)

	api := r.Group("/api")
	{
		accounts := api.Group("/accounts")
		{
			accounts.POST("/verify", accountsHandler.VerifyAccountHandler)
			accounts.POST("/connect", accountsHandler.ConnectAccountHandler)
			accounts.GET("", accountsHandler.ListAccountsHandler)
			accounts.DELETE("/:id", accountsHandler.DisconnectAccountHandler)
		}
	}

	// Start server
	log.Println("Starting CloudCop API server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
