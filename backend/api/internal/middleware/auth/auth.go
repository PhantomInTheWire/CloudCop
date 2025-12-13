// Package auth provides authentication middleware and utilities.
package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/gin-gonic/gin"
)

var userCtxKey = &contextKey{"userId"}

type contextKey struct {
	name string
}

// Middleware verifies the Authorization header (Clerk) and adds the user to context.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request

		// Self-hosted mode mock auth
		selfHosting := os.Getenv("SELF_HOSTING") != ""
		if selfHosting {
			firstName := "Self"
			lastName := "Hosted"
			emailID := "mock_email_id"
			email := "support@cloudcop.dev"
			user := clerk.User{
				ID:                    "mock_user_id",
				FirstName:             &firstName,
				LastName:              &lastName,
				PrimaryEmailAddressID: &emailID,
				EmailAddresses: []clerk.EmailAddress{
					{
						ID:           emailID,
						EmailAddress: email,
					},
				},
			}
			ctx := AttachContext(r.Context(), &user)
			c.Request = r.WithContext(ctx)
			c.Next()
			return
		}

		clientKey := os.Getenv("CLERK_SECRET_KEY")
		// if clientKey == "" {
		// We don't want to panic here in case of misconfiguration in dev, just warn
		// log.Println("WARNING: CLERK_SECRET_KEY is missing")
		// }

		// If no client available (no key), or no header, just finish
		// Ideally we should block if key is present but header missing
		if clientKey == "" {
			c.Next()
			return
		}

		client, err := clerk.NewClient(clientKey)
		if err != nil {
			log.Printf("Failed to create clerk client: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		header := r.Header.Get("Authorization")
		if header == "" {
			// Unauthenticated
			c.Next()
			return
		}

		parts := strings.Split(header, " ")
		if len(parts) != 2 {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		sessionToken := parts[1]

		sessClaims, err := client.VerifyToken(sessionToken)
		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		user, err := client.Users().Read(sessClaims.Subject)
		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		ctx := AttachContext(r.Context(), user)
		c.Request = r.WithContext(ctx)
		c.Next()
	}
}

// AttachContext attaches the user to the context.
func AttachContext(ctx context.Context, user *clerk.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

// FromContext retrieves the user from the context.
func FromContext(ctx context.Context) *clerk.User {
	raw, _ := ctx.Value(userCtxKey).(*clerk.User)
	return raw
}

// EmailFromContext retrieves the primary email from the user in context.
func EmailFromContext(ctx context.Context) (string, error) {
	user := FromContext(ctx)
	if user == nil {
		return "", fmt.Errorf("not logged in")
	}
	// Simplified email retrieval
	if len(user.EmailAddresses) > 0 {
		return user.EmailAddresses[0].EmailAddress, nil
	}
	return "", fmt.Errorf("no email found")
}

// FullnameFromContext retrieves the full name from the user in context.
func FullnameFromContext(ctx context.Context) (string, error) {
	user := FromContext(ctx)
	if user == nil {
		return "", fmt.Errorf("not logged in")
	}
	firstName := ""
	lastName := ""
	if user.FirstName != nil {
		firstName = *user.FirstName
	}
	if user.LastName != nil {
		lastName = *user.LastName
	}
	return fmt.Sprintf("%s %s", firstName, lastName), nil
}
