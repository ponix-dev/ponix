package domain

import (
	"context"
	"errors"
)

type contextKey string

var (
	// ErrMissingUserInContext is returned when a user ID is expected in context but not found.
	ErrMissingUserInContext = errors.New("missing user in context")
)

const (
	// UserKey is the context key for storing the authenticated user ID.
	UserKey contextKey = "user_id"
	// SuperAdminKey is the context key for storing the super admin flag.
	SuperAdminKey contextKey = "super_admin"
)

// SetSuperAdminContext marks the context as belonging to a super admin user.
func SetSuperAdminContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, SuperAdminKey, true)
	return ctx
}

// IsSuperAdminFromContext checks if the context belongs to a super admin user.
func IsSuperAdminFromContext(ctx context.Context) bool {
	isSuperAdmin, ok := ctx.Value(SuperAdminKey).(bool)
	if ok {
		return isSuperAdmin
	}

	return false
}

// SetUserContext adds the authenticated user ID to the context.
func SetUserContext(ctx context.Context, user string) context.Context {
	ctx = context.WithValue(ctx, UserKey, user)
	return ctx
}

// GetUserFromContext extracts the authenticated user ID from context.
// Returns the user ID and true if found, or empty string and false if not found.
func GetUserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(UserKey).(string)
	if ok {
		return user, true
	}

	return "", false
}
