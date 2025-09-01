package connectrpc

import (
	"context"
	"fmt"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type UserManager interface {
	CreateUser(ctx context.Context, createReq *organizationv1.CreateUserRequest) (*organizationv1.User, error)
	GetUser(ctx context.Context, userReq *organizationv1.GetUserRequest) (*organizationv1.User, error)
}

type UserAuthorizer interface {
	CanReadSelf(ctx context.Context, userId, targetUserId string) (bool, error)
	CanUpdateSelf(ctx context.Context, userId, targetUserId string) (bool, error)
}

type UserHandler struct {
	userManager UserManager
	authorizer  UserAuthorizer
}

func NewUserHandler(userManager UserManager, authorizer UserAuthorizer) *UserHandler {
	return &UserHandler{
		userManager: userManager,
		authorizer:  authorizer,
	}
}

func (handler *UserHandler) CreateUser(ctx context.Context, req *connect.Request[organizationv1.CreateUserRequest]) (*connect.Response[organizationv1.CreateUserResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateUser")
	defer span.End()

	// Authorization: CreateUser is super admin only
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user creation is restricted to super admins only"))
	}

	user, err := handler.userManager.CreateUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	response := &organizationv1.CreateUserResponse{
		UserId:    user.GetId(),
		FirstName: user.GetFirstName(),
		LastName:  user.GetLastName(),
		Email:     user.GetEmail(),
		CreatedAt: user.GetCreatedAt(),
	}

	return connect.NewResponse(response), nil
}

func (handler *UserHandler) GetUser(ctx context.Context, req *connect.Request[organizationv1.GetUserRequest]) (*connect.Response[organizationv1.GetUserResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "GetUser")
	defer span.End()

	callingUserId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
	}

	// Authorization: Super admin can get anyone, users can get themselves
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanReadSelf(ctx, callingUserId, req.Msg.GetUserId())
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user %s not authorized to read user %s", callingUserId, req.Msg.GetUserId()))
	}

	user, err := handler.userManager.GetUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	response := &organizationv1.GetUserResponse{
		User: user,
	}

	return connect.NewResponse(response), nil
}
