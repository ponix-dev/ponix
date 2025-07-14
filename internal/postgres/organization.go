package postgres

import (
	"context"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ponix-dev/ponix/internal/postgres/sqlc"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrganizationStore struct {
	db   *sqlc.Queries
	pool *pgxpool.Pool
}

func NewOrganizationStore(db *sqlc.Queries, pool *pgxpool.Pool) *OrganizationStore {
	return &OrganizationStore{
		db:   db,
		pool: pool,
	}
}

func (store *OrganizationStore) CreateOrganization(ctx context.Context, organization *organizationv1.Organization) error {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateOrganization")
	defer span.End()

	params := sqlc.CreateOrganizationParams{
		ID:        organization.GetId(),
		Name:      organization.GetName(),
		Status:    int32(organization.GetStatus()),
		CreatedAt: pgtype.Timestamptz{Time: organization.GetCreatedAt().AsTime(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: organization.GetUpdatedAt().AsTime(), Valid: true},
	}

	_, err := store.db.CreateOrganization(ctx, params)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	return nil
}

func (store *OrganizationStore) GetOrganization(ctx context.Context, organizationID string) (*organizationv1.Organization, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetOrganization")
	defer span.End()

	org, err := store.db.GetOrganization(ctx, organizationID)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	return &organizationv1.Organization{
		Id:        org.ID,
		Name:      org.Name,
		Status:    organizationv1.OrganizationStatus(org.Status),
		CreatedAt: timestamppb.New(org.CreatedAt.Time),
		UpdatedAt: timestamppb.New(org.UpdatedAt.Time),
	}, nil
}
