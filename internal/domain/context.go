package domain

import (
	"context"
	"errors"
)

type contextKey string

var (
	ErrMissingUserInContext = errors.New("missing user in context")
)

const (
	UserKey       contextKey = "user_id"
	SuperAdminKey contextKey = "super_admin"
)

func SetSuperAdminContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, SuperAdminKey, true)
	return ctx
}

func IsSuperAdminFromContext(ctx context.Context) bool {
	isSuperAdmin, ok := ctx.Value(SuperAdminKey).(bool)
	if ok {
		return isSuperAdmin
	}

	return false
}

// SetUserContext adds user information to context
func SetUserContext(ctx context.Context, user string) context.Context {
	ctx = context.WithValue(ctx, UserKey, user)
	return ctx
}

// GetUserFromContext extracts user  from context
func GetUserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(UserKey).(string)
	if ok {
		return user, true
	}

	return "", false
}
