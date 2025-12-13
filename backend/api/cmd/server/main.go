// Package main is the entry point for the CloudCop API server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloudcop/api/internal/awsauth"
	"cloudcop/api/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	auth, err := awsauth.NewAWSAuth()
	if err != nil {
		log.Fatalf("Failed to initialize AWS auth: %v", err)
	}

	cache := awsauth.NewCredentialCache(auth)
	accountsHandler := handlers.NewAccountsHandler(auth, cache)

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

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	/*
		Graceful shutdown: listen for SIGINT/SIGTERM, stop the credential cache
		background goroutine, then shutdown the HTTP server with a timeout.
	*/
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Starting CloudCop API server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	cache.Stop()
	log.Println("Credential cache stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
