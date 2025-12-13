package handlers

import (
	"net/http"

	"cloudcop/api/internal/awsauth"

	"github.com/gin-gonic/gin"
)

// AccountsHandler manages AWS account connection endpoints
type AccountsHandler struct {
	auth  *awsauth.AWSAuth
	cache *awsauth.CredentialCache
}

// NewAccountsHandler creates a new accounts handler
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

// VerifyAccountHandler verifies AWS account access via STS AssumeRole
// POST /api/accounts/verify
func (h *AccountsHandler) VerifyAccountHandler(c *gin.Context) {
	var req VerifyAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Verify access by assuming the role
	accountInfo, err := h.auth.VerifyAccountAccess(c.Request.Context(), awsauth.AssumeRoleInput{
		AccountID:  req.AccountID,
		ExternalID: req.ExternalID,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == awsauth.ErrAssumeRoleFailed || err == awsauth.ErrInvalidExternalID {
			statusCode = http.StatusUnauthorized
		}
		c.JSON(statusCode, gin.H{
			"error":   "Failed to verify AWS account access",
			"details": err.Error(),
		})
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
func (h *AccountsHandler) ConnectAccountHandler(c *gin.Context) {
	var req ConnectAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Verify access first
	accountInfo, err := h.auth.VerifyAccountAccess(c.Request.Context(), awsauth.AssumeRoleInput{
		AccountID:  req.AccountID,
		ExternalID: req.ExternalID,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == awsauth.ErrAssumeRoleFailed || err == awsauth.ErrInvalidExternalID {
			statusCode = http.StatusUnauthorized
		}
		c.JSON(statusCode, gin.H{
			"error":   "Failed to verify AWS account access",
			"details": err.Error(),
		})
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
func (h *AccountsHandler) DisconnectAccountHandler(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Account ID is required",
		})
		return
	}

	// Invalidate cached credentials
	h.cache.InvalidateCredentials(accountID)

	// TODO: Delete from database
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account disconnected successfully",
	})
}
