package connectrpc

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/domain"
)

type SuperAdminer interface {
	IsSuperAdmin(user string) (bool, error)
}

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

type CanAccessEndDevicer interface {
	CanAccessEndDevice(ctx context.Context, userId string, action string, organizationId string) (bool, error)
}

// EndDeviceAuthInterceptor creates an interceptor for end device operations
func EndDeviceAuthInterceptor(enforcer CanAccessEndDevicer) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if domain.IsSuperAdminFromContext(ctx) {
				return next(ctx, req)
			}

			user, ok := domain.GetUserFromContext(ctx)
			if !ok {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
			}

			organization := GetOrganizationFromRequest(req)
			if organization == "" {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("organization  required"))
			}

			action := getActionFromMethod(req.Spec().Procedure)

			allowed, err := enforcer.CanAccessEndDevice(ctx, user, action, organization)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
			}

			if !allowed {
				return nil, connect.NewError(connect.CodePermissionDenied,
					fmt.Errorf("user %s not authorized to %s end devices in organization %s", user, action, organization))
			}

			return next(ctx, req)
		}
	}
}

// TODO: this isn't a good way to do this, but it works for now
// getActionFromMethod maps RPC method names to authorization actions
func getActionFromMethod(procedure string) string {
	method := strings.ToLower(procedure)

	switch {
	case strings.Contains(method, "create"), strings.Contains(method, "add"):
		return "create"
	case strings.Contains(method, "get"), strings.Contains(method, "list"):
		return "read"
	case strings.Contains(method, "update"):
		return "update"
	case strings.Contains(method, "delete"):
		return "delete"
	default:
		return "read" // default to read permission
	}
}

// GetOrganizationFromRequest extracts organization  from the request
// This implementation would depend on your specific request structure
func GetOrganizationFromRequest(req connect.AnyRequest) string {
	// Example implementation - you would adapt this based on your actual request types
	// This might come from:
	// 1. Request message fields
	// 2. Request headers
	// 3. URL path parameters

	// For demonstration, check if there's an org header
	if org := req.Header().Get("X-Organization-"); org != "" {
		return org
	}

	// You might also extract from the request message if it has an organization field
	// This would require type assertions based on your specific protobuf messages

	return ""
}
