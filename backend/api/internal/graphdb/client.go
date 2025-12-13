// Package graphdb provides a client for interacting with the Neo4j graph database.
package graphdb

import (
	"context"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jClient is a wrapper around the Neo4j driver.
type Neo4jClient struct {
	driver neo4j.DriverWithContext
}

// NewNeo4jClient creates a Neo4jClient by reading NEO4J_URI, NEO4J_USERNAME, and
// NEO4J_PASSWORD from the environment and initializing a driver with that data.
// The provided context is ignored. It returns a *Neo4jClient on success or an
// error if driver creation fails.
func NewNeo4jClient(_ context.Context) (*Neo4jClient, error) {
	uri := os.Getenv("NEO4J_URI")
	username := os.Getenv("NEO4J_USERNAME")
	password := os.Getenv("NEO4J_PASSWORD")

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, err
	}

	return &Neo4jClient{driver: driver}, nil
}

// Close closes the Neo4j driver.
func (c *Neo4jClient) Close(ctx context.Context) error {
	return c.driver.Close(ctx)
}

// RunQuery executes a Cypher query with the provided parameters.
func (c *Neo4jClient) RunQuery(ctx context.Context, query string, params map[string]interface{}) (neo4j.ResultWithContext, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		// Explicitly ignore close error as we are deferring it
		_ = session.Close(ctx)
	}()

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}
