package connectrpc

import (
	"context"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type UserManager interface {
	CreateUser(ctx context.Context, createReq *organizationv1.CreateUserRequest) (*organizationv1.User, error)
	GetUser(ctx context.Context, userReq *organizationv1.UserRequest) (*organizationv1.User, error)
}

type UserHandler struct {
	userManager UserManager
}

func NewUserHandler(userManager UserManager) *UserHandler {
	return &UserHandler{
		userManager: userManager,
	}
}

func (handler *UserHandler) CreateUser(ctx context.Context, req *connect.Request[organizationv1.CreateUserRequest]) (*connect.Response[organizationv1.CreateUserResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateUser")
	defer span.End()

	user, err := handler.userManager.CreateUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	response := &organizationv1.CreateUserResponse{
		UserId:         user.GetId(),
		FirstName:      user.GetFirstName(),
		LastName:       user.GetLastName(),
		Email:          user.GetEmail(),
		OrganizationId: user.GetOrganizationId(),
		CreatedAt:      user.GetCreatedAt(),
	}

	return connect.NewResponse(response), nil
}

func (handler *UserHandler) User(ctx context.Context, req *connect.Request[organizationv1.UserRequest]) (*connect.Response[organizationv1.UserResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "User")
	defer span.End()

	user, err := handler.userManager.GetUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	response := &organizationv1.UserResponse{
		User: user,
	}

	return connect.NewResponse(response), nil
}
