package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"buf.build/gen/go/ponix/ponix/connectrpc/go/iot/v1/iotv1connect"
	"buf.build/gen/go/ponix/ponix/connectrpc/go/organization/v1/organizationv1connect"
	"connectrpc.com/connect"
	"connectrpc.com/validate"

	"github.com/ponix-dev/ponix/internal/casbin"
	"github.com/ponix-dev/ponix/internal/clickhouse"
	"github.com/ponix-dev/ponix/internal/conf"
	"github.com/ponix-dev/ponix/internal/connectrpc"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/mux"
	"github.com/ponix-dev/ponix/internal/nats"
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

	cfg, err := conf.GetConfig[conf.AllInOne](ctx)
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

	err = postgres.RunMigrations(ctx, string(curl))
	if err != nil {
		logger.Error("could not run postgres migrations", slog.Any("err", err))
		os.Exit(1)
	}
	dbpool, err := postgres.NewPool(ctx, curl)
	if err != nil {
		logger.Error("could not create db pool", slog.Any("err", err))
		os.Exit(1)
	}

	dbQueries := sqlc.New(dbpool)

	edStore := postgres.NewEndDeviceStore(dbQueries, dbpool)
	orgStore := postgres.NewOrganizationStore(dbQueries, dbpool)
	userStore := postgres.NewUserStore(dbQueries, dbpool)
	userOrgStore := postgres.NewUserOrganizationStore(dbQueries, dbpool)

	pgxAdapter, err := postgres.NewCasbinAdapter(dbpool)
	if err != nil {
		logger.Error("could not create casbin adapter", slog.Any("err", err))
		os.Exit(1)
	}

	casbinEnforcer, err := casbin.NewEnforcer(ctx, pgxAdapter)
	if err != nil {
		logger.Error("could not create casbin enforcer", slog.Any("err", err))
		os.Exit(1)
	}

	// Create domain-specific enforcers
	superAdminEnforcer := casbin.NewSuperAdminEnforcer(casbinEnforcer)
	userEnforcer := casbin.NewUserEnforcer(casbinEnforcer)
	organizationEnforcer := casbin.NewOrganizationEnforcer(casbinEnforcer)
	organizationAccessEnforcer := casbin.NewOrganizationAccessEnforcer(casbinEnforcer)
	endDeviceEnforcer := casbin.NewEndDeviceEnforcer(casbinEnforcer)
	lorawanEnforcer := casbin.NewLoRaWANEnforcer(casbinEnforcer)

	ttnClient, err := ttn.NewTTNClient(
		ttn.WithRegion(ttn.TTNRegion(cfg.TTNRegion)),
		ttn.WithServerName(cfg.TTNServerName),
		ttn.WithCollaboratorApiKey(cfg.TTNApiKey, cfg.TTNApiCollaborator),
	)
	if err != nil {
		logger.Error("could not create ttn client", slog.Any("err", err))
		os.Exit(1)
	}

	churl := clickhouse.NewUrl(cfg.ClickHouseDB, cfg.ClickHouseUser, cfg.ClickHousePass, cfg.ClickHouseAddr)
	err = clickhouse.RunMigrations(ctx, churl)
	if err != nil {
		logger.Error("could not run clickhouse migrations", slog.Any("err", err))
		os.Exit(1)
	}

	clickhouseConn, err := clickhouse.NewConnection(ctx, cfg.ClickHouseDB, cfg.ClickHouseUser, cfg.ClickHousePass, cfg.ClickHouseAddr)
	if err != nil {
		logger.Error("could not create clickhouse connection", slog.Any("err", err))
		os.Exit(1)
	}

	envelopeStore := clickhouse.NewEnvelopeStore(clickhouseConn, cfg.ClickHouseProcessedEnvelopeTable)

	natsConnection, err := nats.NewConnection(
		nats.WithURL(cfg.NatsURL),
		nats.WithName("serviceName"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	jetstreamClient, err := nats.NewJetStream(natsConnection)
	if err != nil {
		log.Fatalf("Failed to create JetStream connection: %v", err)
	}

	//TODO: don't have this do hardcoded setup, use config values
	err = nats.SetupJetStream(ctx, jetstreamClient)
	if err != nil {
		log.Fatalf("Failed to setup JetStream: %v", err)
	}

	processedEnvelopeProducer := nats.NewProcessedEnvelopeProducer(jetstreamClient, cfg.NatsProcessedEnvelopeStream)

	envelopeManager := domain.NewDataEnvelopeManager(processedEnvelopeProducer, envelopeStore, edStore)

	messageHandler := nats.NewProcessedEnvelopeMessageHandler(envelopeManager)
	consumer, err := nats.NewJetStreamConsumer(context.Background(), jetstreamClient, cfg.NatsProcessedEnvelopeStream, "serviceName", cfg.NatsProcessedEnvelopeSubject)
	if err != nil {
		log.Fatalf("Failed to create JetStream consumer: %v", err)
	}

	consumerHandler, err := nats.NewConsumerHandler(consumer, messageHandler, cfg.NatsProcessedEnvelopeBatchSize, cfg.NatsProcessedEnvelopeBatchWait)
	if err != nil {
		log.Fatalf("Failed to create consumer handler: %v", err)
	}

	edMgr := domain.NewEndDeviceManager(edStore, ttnClient, cfg.ApplicationId, xid.StringId, protobuf.Validate)
	edDataMgr := domain.NewEndDeviceDataManager(envelopeStore, protobuf.Validate)
	lorawanMgr := domain.NewLoRaWANManager(edStore, xid.StringId, protobuf.Validate)
	userOrgMgr := domain.NewUserOrganizationManager(userOrgStore, organizationEnforcer, protobuf.Validate)
	organizationManager := domain.NewOrganizationManager(
		orgStore,
		xid.StringId,
		protobuf.Validate,
		userOrgMgr,
	)
	userManager := domain.NewUserManager(
		userStore,
		xid.StringId,
		protobuf.Validate,
	)

	protovalidateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		logger.Error("could not create protovalidate interceptor", slog.Any("err", err))
		os.Exit(1)
	}

	authenticationInterceptor := connectrpc.AuthenticationInterceptor()
	superAdminInterceptor := connectrpc.SuperAdminInterceptor(superAdminEnforcer)

	srv, err := mux.New(
		mux.NewChiMux(),
		mux.WithLogger(logger),
		mux.WithPort(cfg.Port),

		// Organization
		mux.WithHandler(organizationv1connect.NewOrganizationServiceHandler(
			connectrpc.NewOrganizationHandler(organizationManager, organizationAccessEnforcer),
			connect.WithInterceptors(
				authenticationInterceptor,
				superAdminInterceptor,
				protovalidateInterceptor,
			),
		)),
		mux.WithHandler(organizationv1connect.NewUserServiceHandler(
			connectrpc.NewUserHandler(userManager, userEnforcer),
			connect.WithInterceptors(
				authenticationInterceptor,
				superAdminInterceptor,
				protovalidateInterceptor,
			),
		)),
		mux.WithHandler(organizationv1connect.NewOrganizationUserServiceHandler(
			connectrpc.NewOrganizationUserHandler(userOrgMgr, organizationEnforcer),
			connect.WithInterceptors(
				authenticationInterceptor,
				superAdminInterceptor,
				protovalidateInterceptor,
			),
		)),

		// IoT
		mux.WithHandler(iotv1connect.NewEndDeviceServiceHandler(
			connectrpc.NewEndDeviceHandler(edMgr, endDeviceEnforcer),
			connect.WithInterceptors(
				authenticationInterceptor,
				superAdminInterceptor,
				protovalidateInterceptor,
			),
		)),

		mux.WithHandler(iotv1connect.NewLoRaWANServiceHandler(
			connectrpc.NewLoRaWANHandler(lorawanMgr, lorawanEnforcer),
			connect.WithInterceptors(
				authenticationInterceptor,
				superAdminInterceptor,
				protovalidateInterceptor,
			),
		)),

		mux.WithHandler(iotv1connect.NewEndDeviceDataServiceHandler(
			connectrpc.NewEndDeviceDataHandler(edDataMgr, endDeviceEnforcer),
			connect.WithInterceptors(
				authenticationInterceptor,
				superAdminInterceptor,
				protovalidateInterceptor,
			),
		)),

		// Data ingestion handler (no auth interceptors for MVP)
		mux.WithHandler(iotv1connect.NewDataIngestionServiceHandler(
			connectrpc.NewIngestionHandler(envelopeManager),
			connect.WithInterceptors(protovalidateInterceptor),
		)),
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
		runner.WithAppProcess(nats.ConsumerRunner(consumerHandler)),
		runner.WithCloser(telemetry.LoggerProviderCloser()),
		runner.WithCloser(nats.ConnectionCloser(natsConnection)),
	)

	r.Run()
}
