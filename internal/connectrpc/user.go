package connectrpc

import (
	"context"

	organizationv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/organization/v1"
	"connectrpc.com/connect"
	"github.com/ponix-dev/ponix/internal/telemetry"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (handler *UserHandler) CreateUser(ctx context.Context, req *connect.Request[organizationv1.CreateUserRequest]) (*connect.Response[organizationv1.CreateUserResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "CreateUser")
	defer span.End()

	return nil, nil
}

func (handler *UserHandler) User(ctx context.Context, req *connect.Request[organizationv1.UserRequest]) (*connect.Response[organizationv1.UserResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "User")
	defer span.End()

	return nil, nil
}
