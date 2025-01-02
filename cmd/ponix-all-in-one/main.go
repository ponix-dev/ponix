package main

import (
	"context"
	"log/slog"
	"os"

	"buf.build/gen/go/ponix/ponix/connectrpc/go/iot/v1/iotv1connect"
	"buf.build/gen/go/ponix/ponix/connectrpc/go/organization/v1/organizationv1connect"
	"buf.build/gen/go/ponix/ponix/connectrpc/go/ponix/v1/ponixv1connect"

	"github.com/go-chi/chi/v5"
	"github.com/ponix-dev/ponix/internal/connectrpc"
	"github.com/ponix-dev/ponix/internal/mux"
	"github.com/ponix-dev/ponix/internal/runner"
	"github.com/ponix-dev/ponix/internal/telemetry"
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

	srv, err := mux.New(
		mux.NewChiMux(chi.NewRouter()),
		mux.WithLogger(logger),
		mux.WithPort("3000"),

		// System
		mux.WithHandler(ponixv1connect.NewSystemServiceHandler(connectrpc.NewSystemHandler())),
		mux.WithHandler(ponixv1connect.NewSystemInputServiceHandler(connectrpc.NewSystemInputHandler())),

		// Organization
		mux.WithHandler(organizationv1connect.NewOrganizationServiceHandler(connectrpc.NewOrganizationHandler())),
		mux.WithHandler(organizationv1connect.NewUserServiceHandler(connectrpc.NewUserHandler())),

		// IoT
		mux.WithHandler(iotv1connect.NewNetworkServerServiceHandler(connectrpc.NewNetworkServerHandler())),
		mux.WithHandler(iotv1connect.NewGatewayServiceHandler(connectrpc.NewGatewayHandler())),
		mux.WithHandler(iotv1connect.NewEndDeviceServiceHandler(connectrpc.NewEndDeviceHandler())),
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