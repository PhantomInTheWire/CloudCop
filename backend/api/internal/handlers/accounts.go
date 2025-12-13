package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"cloudcop/api/internal/awsauth"
	"cloudcop/api/internal/database"
	"cloudcop/api/internal/middleware/auth"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// AccountsHandler manages AWS account connection endpoints
type AccountsHandler struct {
	auth  *awsauth.AWSAuth
	cache *awsauth.CredentialCache
	store *database.Queries
}

// NewAccountsHandler constructs an AccountsHandler wired with the provided AWS authentication helper,
// credential cache, and database queries. It returns a handler ready to be registered with HTTP routes.
func NewAccountsHandler(auth *awsauth.AWSAuth, cache *awsauth.CredentialCache, store *database.Queries) *AccountsHandler {
	return &AccountsHandler{
		auth:  auth,
		cache: cache,
		store: store,
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
func (h *AccountsHandler) VerifyAccountHandler(c *gin.Context) {
	// Ensure authenticated
	if auth.FromContext(c.Request.Context()) == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
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
func (h *AccountsHandler) ConnectAccountHandler(c *gin.Context) {
	user := auth.FromContext(c.Request.Context())
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
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

	// Ensure user exists in DB
	email, _ := auth.EmailFromContext(c.Request.Context())
	name, _ := auth.FullnameFromContext(c.Request.Context())
	dbUser, err := h.store.CreateUser(c.Request.Context(), database.CreateUserParams{
		ID:    user.ID,
		Email: email,
		Name:  pgtype.Text{String: name, Valid: name != ""},
	})
	if err != nil {
		// Log error but might proceed if user already exists (Query uses ON CONFLICT DO UPDATE)
		log.Printf("Error ensuring user exists: %v", err)
	}

	// Ensure Team exists (MVP: Auto-create team for user if not exists)
	team, err := h.store.GetTeamByOwnerID(c.Request.Context(), user.ID)
	if err != nil {
		// If not found, create
		slug := user.ID // simplified slug
		team, err = h.store.CreateTeam(c.Request.Context(), database.CreateTeamParams{
			Name:    name + "'s Team",
			Slug:    slug,
			OwnerID: user.ID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team context"})
			return
		}
		// Add member
		_, _ = h.store.AddTeamMember(c.Request.Context(), database.AddTeamMemberParams{
			TeamID: team.ID,
			UserID: dbUser.ID,
			Role:   "owner",
		})
	}

	// Store connection in DB
	acct, err := h.store.CreateAccount(c.Request.Context(), database.CreateAccountParams{
		TeamID:         pgtype.Int4{Int32: team.ID, Valid: true},
		AccountID:      accountInfo.AccountID,
		ExternalID:     req.ExternalID, // Use the verified external ID from request
		RoleArn:        pgtype.Text{String: accountInfo.ARN, Valid: true},
		Verified:       pgtype.Bool{Bool: true, Valid: true},
		LastVerifiedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store account connection"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Account connection created successfully",
		"connection": gin.H{
			"id":         acct.ID,
			"account_id": acct.AccountID,
			"arn":        acct.RoleArn.String,
			"verified":   acct.Verified.Bool,
			"user_id":    user.ID,
		},
	})
}

// ListAccountsHandler lists all connected AWS accounts for the authenticated user
// GET /api/accounts
func (h *AccountsHandler) ListAccountsHandler(c *gin.Context) {
	// Retrieve user from context
	user := auth.FromContext(c.Request.Context())
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// First, check if user has a team (MVP: 1 team per user)
	team, err := h.store.GetTeamByOwnerID(c.Request.Context(), user.ID)
	if err != nil {
		// No team, so no accounts
		c.JSON(http.StatusOK, gin.H{
			"accounts": []interface{}{},
		})
		return
	}

	// Get accounts for the team
	accounts, err := h.store.GetAccountsByTeamID(c.Request.Context(), pgtype.Int4{Int32: team.ID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	// Map DB rows to response
	response := make([]gin.H, len(accounts))
	for i, acc := range accounts {
		response[i] = gin.H{
			"id":            acc.ID,
			"account_id":    acc.AccountID,
			"external_id":   acc.ExternalID,
			"role_arn":      acc.RoleArn.String,
			"verified":      acc.Verified.Bool,
			"last_verified": acc.LastVerifiedAt.Time,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": response,
	})
}

// DisconnectAccountHandler removes an AWS account connection
// DELETE /api/accounts/:id
func (h *AccountsHandler) DisconnectAccountHandler(c *gin.Context) {
	accountIDParam := c.Param("id")
	if accountIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Account ID is required",
		})
		return
	}

	user := auth.FromContext(c.Request.Context())
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify ownership via Team
	team, err := h.store.GetTeamByOwnerID(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Team not found"})
		return
	}

	// Find the account first to get ExternalID for cache invalidation
	// We might need a GetAccountByID query, but for now assuming we pass AWS Account ID or DB ID?
	// The route param is :id. Let's assume it's the AWS Account ID string for simplicity in this MVP,
	// or we add a query to get by DB ID.
	// Let's assume it's the AWS Account ID for now strictly.

	// Invalidate credentials
	h.cache.InvalidateCredentials(accountIDParam, "")

	// Delete from DB
	err = h.store.DeleteAccount(c.Request.Context(), database.DeleteAccountParams{
		AccountID: accountIDParam,
		TeamID:    pgtype.Int4{Int32: team.ID, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account disconnected successfully",
	})
}
