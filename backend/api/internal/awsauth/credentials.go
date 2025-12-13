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
	credentials map[string]*cachedCredentials // key: "accountID:externalID"
	auth        *AWSAuth
	stopCh      chan struct{}
}

type cachedCredentials struct {
	creds       *Credentials
	lastRefresh time.Time
}

// cacheKey generates a composite key from accountID and externalID
func cacheKey(accountID, externalID string) string {
	return accountID + ":" + externalID
}

// NewCredentialCache creates a CredentialCache that stores per-account AWS
// credentials and starts a background goroutine that periodically refreshes
// expiring credentials using the provided AWSAuth.
func NewCredentialCache(auth *AWSAuth) *CredentialCache {
	cache := &CredentialCache{
		credentials: make(map[string]*cachedCredentials),
		auth:        auth,
		stopCh:      make(chan struct{}),
	}

	// Start background refresh goroutine
	go cache.refreshLoop()

	return cache
}

// Stop gracefully shuts down the credential cache
func (c *CredentialCache) Stop() {
	close(c.stopCh)
}

// GetCredentials retrieves credentials from cache or fetches new ones
func (c *CredentialCache) GetCredentials(ctx context.Context, accountID, externalID string) (*Credentials, error) {
	key := cacheKey(accountID, externalID)

	c.mu.RLock()
	cached, exists := c.credentials[key]
	c.mu.RUnlock()

	if exists {
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

	key := cacheKey(accountID, externalID)
	c.mu.Lock()
	c.credentials[key] = &cachedCredentials{
		creds:       creds,
		lastRefresh: time.Now(),
	}
	c.mu.Unlock()

	return creds, nil
}

// InvalidateCredentials removes credentials from cache
func (c *CredentialCache) InvalidateCredentials(accountID, externalID string) {
	key := cacheKey(accountID, externalID)
	c.mu.Lock()
	delete(c.credentials, key)
	c.mu.Unlock()
}

// refreshLoop periodically checks and refreshes expiring credentials
func (c *CredentialCache) refreshLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.refreshExpiring()
		case <-c.stopCh:
			return
		}
	}
}

// refreshExpiring refreshes credentials that are about to expire
func (c *CredentialCache) refreshExpiring() {
	c.mu.RLock()
	type expiringCred struct {
		accountID  string
		externalID string
	}
	var expiring []expiringCred
	for key, cached := range c.credentials {
		if time.Until(cached.creds.Expiration) <= refreshBuffer {
			// Parse composite key back to accountID and externalID
			// This is a simple split - in production you might want more robust parsing
			parts := splitCacheKey(key)
			if len(parts) == 2 {
				expiring = append(expiring, expiringCred{
					accountID:  parts[0],
					externalID: parts[1],
				})
			}
		}
	}
	c.mu.RUnlock()

	// Refresh expiring credentials
	for _, cred := range expiring {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if _, err := c.RefreshCredentials(ctx, cred.accountID, cred.externalID); err != nil {
			// TODO: Add proper logging when logger is available
			// For now, silently continue to avoid crashes
			_ = err
		}
		cancel()
	}
}

// splitCacheKey splits a composite cache key into accountID and externalID
func splitCacheKey(key string) []string {
	// Find the first colon to split accountID and externalID
	for i := 0; i < len(key); i++ {
		if key[i] == ':' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return []string{key}
}

// GetCachedCredentialsCount returns the number of cached credentials
func (c *CredentialCache) GetCachedCredentialsCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.credentials)
}
