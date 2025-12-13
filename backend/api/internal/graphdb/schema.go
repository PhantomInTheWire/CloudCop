package graphdb

import (
	"context"
	"fmt"
)

// InitializeSchema creates necessary constraints and indexes in Neo4j.
func (c *Neo4jClient) InitializeSchema(ctx context.Context) error {
	queries := []string{
		"CREATE CONSTRAINT IF NOT EXISTS FOR (n:EC2Instance) REQUIRE n.instance_id IS UNIQUE",
		"CREATE INDEX IF NOT EXISTS FOR (n:EC2Instance) ON (n.instance_id)",
		"CREATE INDEX IF NOT EXISTS FOR (n:S3Bucket) ON (n.bucket_name)",
		"CREATE INDEX IF NOT EXISTS FOR (n:IAMRole) ON (n.role_name)",
	}

	for _, query := range queries {
		_, err := c.RunQuery(ctx, query, nil)
		if err != nil {
			return fmt.Errorf("failed to execute schema query '%s': %w", query, err)
		}
	}
	return nil
}
