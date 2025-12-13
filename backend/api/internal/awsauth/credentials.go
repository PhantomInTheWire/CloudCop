package awsauth

import (
	"context"
	"sync"
	"time"
)

const (
	// Refresh credentials 5 minutes before expiration
	refreshBuffer = 5 * time.Minute
)

// CredentialCache manages cached AWS credentials with automatic refresh
type CredentialCache struct {
	mu          sync.RWMutex
	credentials map[string]*cachedCredentials
	auth        *AWSAuth
}

type cachedCredentials struct {
	creds       *Credentials
	externalID  string
	lastRefresh time.Time
}

// NewCredentialCache creates a CredentialCache that stores per-account AWS
// credentials and starts a background goroutine that periodically refreshes
// expiring credentials using the provided AWSAuth.
func NewCredentialCache(auth *AWSAuth) *CredentialCache {
	cache := &CredentialCache{
		credentials: make(map[string]*cachedCredentials),
		auth:        auth,
	}

	// Start background refresh goroutine
	go cache.refreshLoop()

	return cache
}

// GetCredentials retrieves credentials from cache or fetches new ones
func (c *CredentialCache) GetCredentials(ctx context.Context, accountID, externalID string) (*Credentials, error) {
	c.mu.RLock()
	cached, exists := c.credentials[accountID]
	c.mu.RUnlock()

	if exists && cached.externalID == externalID {
		// Check if credentials are still valid
		if time.Until(cached.creds.Expiration) > refreshBuffer {
			return cached.creds, nil
		}
	}

	// Fetch new credentials
	return c.RefreshCredentials(ctx, accountID, externalID)
}

// RefreshCredentials fetches new credentials and updates the cache
func (c *CredentialCache) RefreshCredentials(ctx context.Context, accountID, externalID string) (*Credentials, error) {
	creds, err := c.auth.AssumeRole(ctx, AssumeRoleInput{
		AccountID:  accountID,
		ExternalID: externalID,
	})
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.credentials[accountID] = &cachedCredentials{
		creds:       creds,
		externalID:  externalID,
		lastRefresh: time.Now(),
	}
	c.mu.Unlock()

	return creds, nil
}

// InvalidateCredentials removes credentials from cache
func (c *CredentialCache) InvalidateCredentials(accountID string) {
	c.mu.Lock()
	delete(c.credentials, accountID)
	c.mu.Unlock()
}

// refreshLoop periodically checks and refreshes expiring credentials
func (c *CredentialCache) refreshLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.refreshExpiring()
	}
}

// refreshExpiring refreshes credentials that are about to expire
func (c *CredentialCache) refreshExpiring() {
	c.mu.RLock()
	expiring := make(map[string]string) // accountID -> externalID
	for accountID, cached := range c.credentials {
		if time.Until(cached.creds.Expiration) <= refreshBuffer {
			expiring[accountID] = cached.externalID
		}
	}
	c.mu.RUnlock()

	// Refresh expiring credentials
	for accountID, externalID := range expiring {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, _ = c.RefreshCredentials(ctx, accountID, externalID)
		cancel()
	}
}

// GetCachedCredentialsCount returns the number of cached credentials
func (c *CredentialCache) GetCachedCredentialsCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.credentials)
}