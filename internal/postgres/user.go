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

// UserStore handles database operations for users.
type UserStore struct {
	db   *sqlc.Queries
	pool *pgxpool.Pool
}

// NewUserStore creates a new UserStore instance.
func NewUserStore(db *sqlc.Queries, pool *pgxpool.Pool) *UserStore {
	return &UserStore{
		db:   db,
		pool: pool,
	}
}

// CreateUser inserts a new user into the database.
func (store *UserStore) CreateUser(ctx context.Context, user *organizationv1.User) error {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateUser")
	defer span.End()

	params := sqlc.CreateUserParams{
		ID:        user.GetId(),
		FirstName: user.GetFirstName(),
		LastName:  user.GetLastName(),
		Email:     user.GetEmail(),
		CreatedAt: pgtype.Timestamptz{Time: user.GetCreatedAt().AsTime(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: user.GetUpdatedAt().AsTime(), Valid: true},
	}

	_, err := store.db.CreateUser(ctx, params)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	return nil
}

// GetUser retrieves a user by ID from the database.
func (store *UserStore) GetUser(ctx context.Context, userID string) (*organizationv1.User, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetUser")
	defer span.End()

	user, err := store.db.GetUser(ctx, userID)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	return &organizationv1.User{
		Id:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt.Time),
		UpdatedAt: timestamppb.New(user.UpdatedAt.Time),
	}, nil
}
