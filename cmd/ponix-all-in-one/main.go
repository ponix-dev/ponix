package main

import (
	"context"
	"log/slog"
	"os"

	"buf.build/gen/go/ponix/ponix/connectrpc/go/iot/v1/iotv1connect"
	"buf.build/gen/go/ponix/ponix/connectrpc/go/organization/v1/organizationv1connect"
	"buf.build/gen/go/ponix/ponix/connectrpc/go/ponix/v1/ponixv1connect"
	"connectrpc.com/connect"
	"connectrpc.com/validate"

	"github.com/go-chi/chi/v5"
	"github.com/ponix-dev/ponix/internal/connectrpc"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/mux"
	"github.com/ponix-dev/ponix/internal/runner"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/xid"
)

var (
	serviceName = "ponix-all-in-one"
)

func main() {
	logger := slog.Default()
	ctx := context.Background()

	resource, err := telemetry.NewResource(serviceName)
	if err != nil {
		logger.Error("could not create resource", slog.Any("err", err))
		os.Exit(1)
	}

	meterProvider, err := telemetry.NewMeterProvider(ctx, resource)
	if err != nil {
		logger.Error("could not create meter provider", slog.Any("err", err))
		os.Exit(1)
	}

	telemetry.SetServiceMeter(meterProvider)

	tracerProvider, err := telemetry.NewTracerProvider(ctx, resource)
	if err != nil {
		logger.Error("could not create tracer provider", slog.Any("err", err))
		os.Exit(1)
	}

	telemetry.SetServiceTracer(tracerProvider)

	systemMgr := domain.NewSystemManager(nil, xid.StringId)
	systemInputMgr := domain.NewSystemInputManager(nil, xid.StringId)
	nsMgr := domain.NewNetworkServerManager(nil, xid.StringId)
	gatewayMgr := domain.NewGatewayManager(nil, xid.StringId)
	edMgr := domain.NewEndDeviceManager(nil, xid.StringId)

	protovalidateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		logger.Error("could not create protovalidate interceptor", slog.Any("err", err))
		os.Exit(1)
	}

	srv, err := mux.New(
		mux.NewChiMux(chi.NewRouter()),
		mux.WithLogger(logger),
		mux.WithPort("3000"),

		// System
		mux.WithHandler(ponixv1connect.NewSystemServiceHandler(
			connectrpc.NewSystemHandler(
				systemMgr,
				nsMgr,
				gatewayMgr,
				edMgr,
				systemInputMgr,
			),
			connect.WithInterceptors(protovalidateInterceptor),
		)),
		mux.WithHandler(ponixv1connect.NewSystemInputServiceHandler(connectrpc.NewSystemInputHandler(systemInputMgr), connect.WithInterceptors(protovalidateInterceptor))),

		// Organization
		mux.WithHandler(organizationv1connect.NewOrganizationServiceHandler(connectrpc.NewOrganizationHandler(), connect.WithInterceptors(protovalidateInterceptor))),
		mux.WithHandler(organizationv1connect.NewUserServiceHandler(connectrpc.NewUserHandler(), connect.WithInterceptors(protovalidateInterceptor))),

		// IoT
		mux.WithHandler(iotv1connect.NewNetworkServerServiceHandler(connectrpc.NewNetworkServerHandler(nsMgr), connect.WithInterceptors(protovalidateInterceptor))),
		mux.WithHandler(iotv1connect.NewGatewayServiceHandler(connectrpc.NewGatewayHandler(gatewayMgr), connect.WithInterceptors(protovalidateInterceptor))),
		mux.WithHandler(iotv1connect.NewEndDeviceServiceHandler(connectrpc.NewEndDeviceHandler(edMgr), connect.WithInterceptors(protovalidateInterceptor))),
	)
	if err != nil {
		logger.Error("could not create server", slog.Any("err", err))
		os.Exit(1)
	}

	r := runner.New(
		runner.WithLogger(logger),
		runner.WithAppProcess(mux.NewRunner(srv)),
		runner.WithCloser(mux.NewCloser(srv)),
		runner.WithCloser(telemetry.MeterProviderCloser(meterProvider)),
		runner.WithCloser(telemetry.TracerProviderCloser(tracerProvider)),
	)

	r.Run()
}
