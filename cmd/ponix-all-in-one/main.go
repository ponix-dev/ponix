package main

import (
	"context"
	"log/slog"
	"os"

	"buf.build/gen/go/ponix/ponix/connectrpc/go/iot/v1/iotv1connect"
	"buf.build/gen/go/ponix/ponix/connectrpc/go/organization/v1/organizationv1connect"
	"connectrpc.com/connect"
	"connectrpc.com/validate"

	"github.com/go-chi/chi/v5"
	"github.com/ponix-dev/ponix/internal/conf"
	"github.com/ponix-dev/ponix/internal/connectrpc"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/mux"
	"github.com/ponix-dev/ponix/internal/postgres"
	"github.com/ponix-dev/ponix/internal/postgres/sqlc"
	"github.com/ponix-dev/ponix/internal/protobuf"
	"github.com/ponix-dev/ponix/internal/runner"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/ttn"
	"github.com/ponix-dev/ponix/internal/xid"
)

var (
	serviceName = "ponix-all-in-one"
)

func main() {
	logger := slog.Default()
	ctx := context.Background()

	cfg, err := conf.GetConfig[conf.ManagementConfig](ctx)
	if err != nil {
		logger.Error("could not get config", slog.Any("err", err))
		os.Exit(1)
	}

	resource, err := telemetry.NewResource(serviceName)
	if err != nil {
		logger.Error("could not create resource", slog.Any("err", err))
		os.Exit(1)
	}

	logger, err = telemetry.NewLogger(ctx, resource, serviceName)
	if err != nil {
		logger.Error("could not create logger", slog.Any("err", err))
		os.Exit(1)
	}

	telemetry.SetLogger(logger)

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

	curl := postgres.NewConnUrl(
		postgres.WithDB(cfg.Database),
		postgres.WithUrl(cfg.DatabaseUrl),
		postgres.WithUser(cfg.DatabaseUsername),
		postgres.WithPassword(cfg.DatabasePassword),
	)

	dbpool, err := postgres.NewPool(ctx, curl)
	if err != nil {
		logger.Error("could not create db pool", slog.Any("err", err))
		os.Exit(1)
	}

	dbQueries := sqlc.New(dbpool)

	edStore := postgres.NewEndDeviceStore(dbQueries, dbpool)

	ttnClient, err := ttn.NewTTNClient(
		ttn.WithRegion(ttn.TTNRegion(cfg.TTNRegion)),
		ttn.WithServerName(cfg.TTNServerName),
		ttn.WithCollaboratorApiKey(cfg.TTNApiKey, cfg.TTNApiCollaborator),
	)
	if err != nil {
		logger.Error("could not create ttn client", slog.Any("err", err))
		os.Exit(1)
	}

	edMgr := domain.NewEndDeviceManager(edStore, ttnClient, cfg.ApplicationId, xid.StringId, protobuf.Validate)
	lorawanMgr := domain.NewLoRaWANManager(edStore, xid.StringId, protobuf.Validate)

	protovalidateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		logger.Error("could not create protovalidate interceptor", slog.Any("err", err))
		os.Exit(1)
	}

	srv, err := mux.New(
		mux.NewChiMux(chi.NewRouter()),
		mux.WithLogger(logger),
		mux.WithPort(cfg.Port),

		// Organization
		mux.WithHandler(organizationv1connect.NewOrganizationServiceHandler(connectrpc.NewOrganizationHandler(), connect.WithInterceptors(protovalidateInterceptor))),
		mux.WithHandler(organizationv1connect.NewUserServiceHandler(connectrpc.NewUserHandler(), connect.WithInterceptors(protovalidateInterceptor))),

		// IoT
		mux.WithHandler(iotv1connect.NewEndDeviceServiceHandler(connectrpc.NewEndDeviceHandler(edMgr), connect.WithInterceptors(protovalidateInterceptor))),
		mux.WithHandler(iotv1connect.NewLoRaWANServiceHandler(connectrpc.NewLoRaWANHandler(lorawanMgr), connect.WithInterceptors(protovalidateInterceptor))),
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
		runner.WithCloser(telemetry.LoggerProviderCloser()),
	)

	r.Run()
}
