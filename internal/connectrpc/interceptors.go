package connectrpc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/domain"
)

// SuperAdminer checks whether a user has super admin privileges.
type SuperAdminer interface {
	IsSuperAdmin(user string) (bool, error)
}

// AuthenticationInterceptor creates an authentication interceptor
// For development, it sets a hardcoded user ID in the context for all requests
// TODO: Replace with real JWT/session validation in production
func AuthenticationInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Set hardcoded user ID in context for all requests (development only)
			ctx = domain.SetUserContext(ctx, "dev-user-123")
			return next(ctx, req)
		}
	}
}

// SuperAdminInterceptor creates an interceptor that checks if the authenticated user has super admin privileges.
// If the user is a super admin, it enriches the request context with super admin status.
// This enables handlers to bypass organization-level authorization checks for administrative operations.
func SuperAdminInterceptor(enforcer SuperAdminer) connect.UnaryInterceptorFunc {
	return func(uf connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Extract user from context (typically set by authentication middleware)
			userId, ok := domain.GetUserFromContext(ctx)
			if !ok {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
			}

			// Check if user is super admin first
			isSuperAdmin, err := enforcer.IsSuperAdmin(userId)
			if err != nil {
				return nil, err
			}

			if isSuperAdmin {
				ctx = domain.SetSuperAdminContext(ctx)
			}

			return uf(ctx, req)
		}
	}
}

// GetOrganizationFromRequest extracts the organization ID from RPC request messages.
// It attempts to extract the organization ID from common field names (OrganizationId, Organization).
// Returns an empty string if the organization ID cannot be extracted from the request.
func GetOrganizationFromRequest(req any) string {
	// Try to extract from the request message
	switch msg := req.(type) {
	case interface{ GetOrganizationId() string }:
		return msg.GetOrganizationId()
	case interface{ GetOrganization() string }:
		return msg.GetOrganization()
	default:
		// Could not extract organization ID from request
		return ""
	}
}