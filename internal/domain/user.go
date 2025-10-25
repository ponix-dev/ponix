package domain

import (
	"context"
	"time"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserStorer defines the persistence operations for users.
type UserStorer interface {
	CreateUser(ctx context.Context, user *organizationv1.User) error
	GetUser(ctx context.Context, userId string) (*organizationv1.User, error)
}

// UserManager orchestrates user-related business logic.
type UserManager struct {
	userStore UserStorer
	stringId  StringId
	validate  Validate
}

// NewUserManager creates a new instance of UserManager with the provided dependencies.
func NewUserManager(us UserStorer, stringId StringId, validate Validate) *UserManager {
	return &UserManager{
		userStore: us,
		stringId:  stringId,
		validate:  validate,
	}
}

// CreateUser creates a new user with a unique ID.
func (mgr *UserManager) CreateUser(ctx context.Context, createReq *organizationv1.CreateUserRequest) (*organizationv1.User, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateUser")
	defer span.End()

	err := mgr.validate(createReq)
	if err != nil {
		return nil, err
	}

	userId := mgr.stringId()

	now := timestamppb.New(time.Now().UTC())

	user := organizationv1.User_builder{
		Id:        userId,
		FirstName: createReq.GetFirstName(),
		LastName:  createReq.GetLastName(),
		Email:     createReq.GetEmail(),
		CreatedAt: now,
		UpdatedAt: now,
	}.Build()

	err = mgr.userStore.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves a user by their ID.
func (mgr *UserManager) GetUser(ctx context.Context, userReq *organizationv1.GetUserRequest) (*organizationv1.User, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetUser")
	defer span.End()

	err := mgr.validate(userReq)
	if err != nil {
		return nil, err
	}

	user, err := mgr.userStore.GetUser(ctx, userReq.GetUserId())
	if err != nil {
		return nil, err
	}

	return user, nil
}
