package handlers

import (
	"context"
	"net/http"

	"cloudcop/api/internal/awsauth"

	"github.com/gin-gonic/gin"
)

// AccountsHandler manages AWS account connection endpoints
type AccountsHandler struct {
	auth  *awsauth.AWSAuth
	cache *awsauth.CredentialCache
}

// NewAccountsHandler creates an AccountsHandler wired with the provided AWS authentication and credential cache.
// The auth parameter supplies AWS authentication/verification functionality; cache is used for storing and invalidating temporary credentials.
// TODO: Add authentication middleware to verify user identity before allowing account operations
func NewAccountsHandler(auth *awsauth.AWSAuth, cache *awsauth.CredentialCache) *AccountsHandler {
	return &AccountsHandler{
		auth:  auth,
		cache: cache,
	}
}

// VerifyAccountRequest represents the request to verify AWS account access
type VerifyAccountRequest struct {
	AccountID  string `json:"account_id" binding:"required"`
	ExternalID string `json:"external_id" binding:"required"`
}

// ConnectAccountRequest represents the request to connect an AWS account
type ConnectAccountRequest struct {
	AccountID  string `json:"account_id" binding:"required"`
	ExternalID string `json:"external_id" binding:"required"`
}

// verifyAccount is a helper that verifies AWS account access
func (h *AccountsHandler) verifyAccount(ctx context.Context, accountID, externalID string) (*awsauth.AccountInfo, error) {
	return h.auth.VerifyAccountAccess(ctx, awsauth.AssumeRoleInput{
		AccountID:  accountID,
		ExternalID: externalID,
	})
}

// handleVerificationError returns appropriate error response for verification failures
func handleVerificationError(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	if err == awsauth.ErrAssumeRoleFailed || err == awsauth.ErrInvalidExternalID {
		statusCode = http.StatusUnauthorized
	}
	c.JSON(statusCode, gin.H{
		"error": "Failed to verify AWS account access",
	})
}

// VerifyAccountHandler verifies AWS account access via STS AssumeRole
// POST /api/accounts/verify
// TODO: Add authentication middleware to verify user is logged in
func (h *AccountsHandler) VerifyAccountHandler(c *gin.Context) {
	var req VerifyAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	// Verify access by assuming the role
	accountInfo, err := h.verifyAccount(c.Request.Context(), req.AccountID, req.ExternalID)
	if err != nil {
		handleVerificationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verified": true,
		"account_info": gin.H{
			"account_id": accountInfo.AccountID,
			"arn":        accountInfo.ARN,
			"user_id":    accountInfo.UserID,
		},
	})
}

// ConnectAccountHandler creates a new AWS account connection
// POST /api/accounts/connect
// TODO: Add authentication middleware to get user ID from context
// TODO: Store user_id with account connection in database for authorization
func (h *AccountsHandler) ConnectAccountHandler(c *gin.Context) {
	var req ConnectAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	// Verify access first
	accountInfo, err := h.verifyAccount(c.Request.Context(), req.AccountID, req.ExternalID)
	if err != nil {
		handleVerificationError(c, err)
		return
	}

	// TODO: Store connection in database
	// For now, just return success
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Account connection created successfully",
		"connection": gin.H{
			"account_id": accountInfo.AccountID,
			"arn":        accountInfo.ARN,
			"verified":   true,
		},
	})
}

// ListAccountsHandler lists all connected AWS accounts for the authenticated user
// GET /api/accounts
func (h *AccountsHandler) ListAccountsHandler(c *gin.Context) {
	// TODO: Retrieve from database
	// For now, return empty list
	c.JSON(http.StatusOK, gin.H{
		"accounts": []interface{}{},
	})
}

// DisconnectAccountHandler removes an AWS account connection
// DELETE /api/accounts/:id
// TODO: Add authorization check to verify the account belongs to the authenticated user
func (h *AccountsHandler) DisconnectAccountHandler(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Account ID is required",
		})
		return
	}

	// Invalidate cached credentials
	// TODO: Track externalID in database to properly invalidate specific credentials
	h.cache.InvalidateCredentials(accountID, "")

	// TODO: Delete from database
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account disconnected successfully",
	})
}
