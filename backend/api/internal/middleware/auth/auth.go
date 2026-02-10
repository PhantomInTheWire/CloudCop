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

// Middleware returns a Gin handler that authenticates requests and, when successful,
// attaches the authenticated Clerk user to the request context.
//
// The handler authenticates using Clerk when CLERK_SECRET_KEY is set; if the key is
// empty or the Authorization header is missing, the request proceeds without a user
// attached. In self-hosted mode (SELF_HOSTING set) a mock user is attached. If Clerk
// client creation fails the handler responds with HTTP 500. Invalid or malformed
// authorization tokens and failures to read the user result in HTTP 403.
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

// AttachContext returns a copy of ctx that carries the provided Clerk user value under the package's user context key.
// The returned context can be used to retrieve the user later; the user argument may be nil.
func AttachContext(ctx context.Context, user *clerk.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

// FromContext retrieves the Clerk user stored in ctx.
// It returns nil if no user has been attached to the context.
func FromContext(ctx context.Context) *clerk.User {
	raw, _ := ctx.Value(userCtxKey).(*clerk.User)
	return raw
}

// EmailFromContext retrieves the email address from the authenticated user in the context.
// If no user is attached to the context or the user has no email addresses an error is returned.
func EmailFromContext(ctx context.Context) (string, error) {
	user := FromContext(ctx)
	if user == nil {
		return "", fmt.Errorf("not logged in")
	}
	if len(user.EmailAddresses) > 0 {
		return user.EmailAddresses[0].EmailAddress, nil
	}
	return "", fmt.Errorf("no email found")
}

// FullnameFromContext returns the user's full name composed of first and last name from the context.
// If no user is attached to the context it returns an error "not logged in". Nil first or last name values are treated as empty strings; the result is formatted as "First Last".
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
