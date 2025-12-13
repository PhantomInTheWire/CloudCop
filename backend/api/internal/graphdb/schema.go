package graphdb

import (
	"context"
	"fmt"
)

// InitializeSchema creates necessary constraints and indexes in Neo4j.
func (c *Neo4jClient) InitializeSchema(ctx context.Context) error {
	queries := []string{
		"CREATE CONSTRAINT IF NOT EXISTS FOR (n:EC2Instance) REQUIRE n.instance_id IS UNIQUE",
		"CREATE INDEX IF NOT EXISTS FOR (n:EC2Instance) ON (n.region)",
		"CREATE CONSTRAINT IF NOT EXISTS FOR (n:S3Bucket) REQUIRE n.name IS UNIQUE",
		"CREATE CONSTRAINT IF NOT EXISTS FOR (n:IAMRole) REQUIRE n.arn IS UNIQUE",
	}

	for _, q := range queries {
		_, err := c.RunQuery(ctx, q, nil)
		if err != nil {
			return fmt.Errorf("failed to execute query '%s': %w", q, err)
		}
	}

	return nil
}
