// Package graph provides the GraphQL resolvers and schema definitions.
package graph

import (
	"cloudcop/api/internal/awsauth"
	"cloudcop/api/internal/database"
	"cloudcop/api/internal/graphdb"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

// Resolver is the dependency injection struct for the graph resolver.
type Resolver struct {
	DB    *database.Queries
	Auth  *awsauth.AWSAuth
	Cache *awsauth.CredentialCache
	Neo4j *graphdb.Neo4jClient
}
