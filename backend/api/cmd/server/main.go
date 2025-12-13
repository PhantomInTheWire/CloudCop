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

	"cloudcop/api/graph"
	"cloudcop/api/internal/awsauth"
	"cloudcop/api/internal/database"
	"cloudcop/api/internal/graphdb"
	"cloudcop/api/internal/handlers"
	"cloudcop/api/internal/middleware/auth"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// main initializes services (PostgreSQL, optional Neo4j, and AWS auth), registers HTTP and GraphQL routes,
// starts the API server on :8080, and performs a graceful shutdown on SIGINT/SIGTERM by stopping the credential
// cache and closing Neo4j and database connections.
//
// If Neo4j initialization fails, the server continues to start without Neo4j support.
func main() {
	// Initialize PostgreSQL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/cloudcop?sslmode=disable"
	}
	connPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	store := database.New(connPool)

	// Initialize Neo4j
	neo4jClient, err := graphdb.NewNeo4jClient(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to initialize Neo4j client: %v", err)
		// Proceeding without Neo4j for now to allow server start if Neo4j is down
	}

	awsAuth, err := awsauth.NewAWSAuth()
	if err != nil {
		log.Fatalf("Failed to initialize AWS auth: %v", err)
	}

	cache := awsauth.NewCredentialCache(awsAuth)
	accountsHandler := handlers.NewAccountsHandler(awsAuth, cache, store)

	r := gin.Default()
	r.GET("/health", handlers.Health)

	api := r.Group("/api")
	api.Use(auth.Middleware()) // Apply auth middleware to API routes including GraphQL
	{
		accounts := api.Group("/accounts")
		{
			accounts.POST("/verify", accountsHandler.VerifyAccountHandler)
			accounts.POST("/connect", accountsHandler.ConnectAccountHandler)
			accounts.GET("", accountsHandler.ListAccountsHandler)
			accounts.DELETE("/:id", accountsHandler.DisconnectAccountHandler)
		}

		// GraphQL Endpoint
		srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
			DB:    store,
			Auth:  awsAuth,
			Cache: cache,
			Neo4j: neo4jClient,
		}}))

		api.POST("/query", func(c *gin.Context) {
			srv.ServeHTTP(c.Writer, c.Request)
		})

		// GraphQL Playground
		api.GET("/playground", func(c *gin.Context) {
			playground.Handler("GraphQL", "/api/query").ServeHTTP(c.Writer, c.Request)
		})
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
	if neo4jClient != nil {
		if err := neo4jClient.Close(context.Background()); err != nil {
			log.Printf("Error closing Neo4j client: %v", err)
		}
	} else {
		log.Println("Neo4j client was not initialized, skipping close")
	}
	connPool.Close()
	log.Println("Credential cache, Neo4j, and DB connections stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
