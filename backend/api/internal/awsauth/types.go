package awsauth

import (
	"errors"
	"time"
)

// Credentials represents temporary AWS credentials from STS AssumeRole
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

// AccountInfo contains AWS account identity information
type AccountInfo struct {
	AccountID string
	ARN       string
	UserID    string
}

// AssumeRoleInput contains parameters for STS AssumeRole
type AssumeRoleInput struct {
	AccountID  string
	ExternalID string
}

// Common error types
var (
	ErrInvalidExternalID   = errors.New("invalid external ID")
	ErrAssumeRoleFailed    = errors.New("failed to assume role")
	ErrCredentialsExpired  = errors.New("credentials have expired")
	ErrSelfHostedMode      = errors.New("operation not supported in self-hosted mode")
	ErrInvalidCredentials  = errors.New("invalid AWS credentials")
)
