package postgres

import (
	"context"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ponix-dev/ponix/internal/postgres/sqlc"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

type UserOrganizationStore struct {
	queries *sqlc.Queries
	db      *pgxpool.Pool
}

func NewUserOrganizationStore(queries *sqlc.Queries, db *pgxpool.Pool) *UserOrganizationStore {
	return &UserOrganizationStore{
		queries: queries,
		db:      db,
	}
}

func (uos *UserOrganizationStore) AddUserToOrganization(ctx context.Context, orgUser *organizationv1.OrganizationUser) error {
	ctx, span := telemetry.Tracer().Start(ctx, "AddUserToOrganization")
	defer span.End()

	err := uos.queries.AddUserToOrganization(ctx, sqlc.AddUserToOrganizationParams{
		UserID:         orgUser.UserId,
		OrganizationID: orgUser.OrganizationId,
		Role:           orgUser.Role,
	})
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to add user to organization: %w", err)
	}

	return nil
}

func (uos *UserOrganizationStore) RemoveUserFromOrganization(ctx context.Context, userId, organizationId string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "RemoveUserFromOrganization")
	defer span.End()

	err := uos.queries.RemoveUserFromOrganization(ctx, sqlc.RemoveUserFromOrganizationParams{
		UserID:         userId,
		OrganizationID: organizationId,
	})
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to remove user from organization: %w", err)
	}
	return nil
}

func (uos *UserOrganizationStore) GetUserOrganizations(ctx context.Context, userId string) ([]*organizationv1.OrganizationUser, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetUserOrganizations")
	defer span.End()

	dbOrgUsers, err := uos.queries.GetUserOrganizations(ctx, userId)
	if err != nil {
		return nil, stacktrace.NewStackTraceErrorf("failed to get user organizations: %w", err)
	}

	orgUsers := make([]*organizationv1.OrganizationUser, len(dbOrgUsers))
	for i, dbOrgUser := range dbOrgUsers {
		orgUsers[i] = &organizationv1.OrganizationUser{
			UserId:         userId,
			OrganizationId: dbOrgUser.OrganizationID,
			Role:           dbOrgUser.Role,
		}
	}

	return orgUsers, nil
}

func (uos *UserOrganizationStore) GetOrganizationUsers(ctx context.Context, organizationID string) ([]*organizationv1.OrganizationUser, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetOrganizationUsers")
	defer span.End()

	dbOrgUsers, err := uos.queries.GetOrganizationUsers(ctx, organizationID)
	if err != nil {
		return nil, stacktrace.NewStackTraceErrorf("failed to get organization users: %w", err)
	}

	orgUsers := make([]*organizationv1.OrganizationUser, len(dbOrgUsers))
	for i, dbOrgUser := range dbOrgUsers {
		orgUsers[i] = &organizationv1.OrganizationUser{
			UserId:         dbOrgUser.UserID,
			OrganizationId: organizationID,
			Role:           dbOrgUser.Role,
		}
	}

	return orgUsers, nil
}

func (uos *UserOrganizationStore) IsUserInOrganization(ctx context.Context, userId, organizationId string) (bool, error) {
	isMember, err := uos.queries.IsUserInOrganization(ctx, sqlc.IsUserInOrganizationParams{
		UserID:         userId,
		OrganizationID: organizationId,
	})
	if err != nil {
		return false, stacktrace.NewStackTraceErrorf("failed to check user organization membership: %w", err)
	}

	return isMember, nil
}

func (uos *UserOrganizationStore) UpdateUserRole(ctx context.Context, userId, organizationId, role string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "UpdateUserRole")
	defer span.End()

	err := uos.queries.UpdateUserRole(ctx, sqlc.UpdateUserRoleParams{
		UserID:         userId,
		OrganizationID: organizationId,
		Role:           role,
	})
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to update user role: %w", err)
	}

	return nil
}
