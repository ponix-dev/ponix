package management_service_test

import (
	"context"
	"net"
	"testing"

	ponixv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/ponix/v1"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/postgres"
	"github.com/ponix-dev/ponix/internal/postgres/sqlc"
	"github.com/ponix-dev/ponix/internal/protobuf"
	"github.com/ponix-dev/ponix/internal/xid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tcp "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestIntegrationSystem(t *testing.T) {
	t.Run("tests system and system entities creation", func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}

		ctx := context.Background()

		dbc, dbQueries := setupDatabase(t, ctx)
		defer testcontainers.CleanupContainer(t, dbc)

		systemStore := postgres.NewSystemStore(dbQueries)
		// systemInputStore := postgres.NewSystemInputStore(dbQueries)
		// nsStore := postgres.NewNetworkServerStore(dbQueries)
		// gatewayStore := postgres.NewGatewayStore(dbQueries)
		// edStore := postgres.NewEndDeviceStore(dbQueries)

		systemMgr := domain.NewSystemManager(systemStore, xid.StringId, protobuf.Validate)
		// systemInputMgr := domain.NewSystemInputManager(systemInputStore, xid.StringId, protobuf.Validate)
		// nsMgr := domain.NewNetworkServerManager(nsStore, xid.StringId, protobuf.Validate)
		// gatewayMgr := domain.NewGatewayManager(gatewayStore, xid.StringId, protobuf.Validate)
		// edMgr := domain.NewEndDeviceManager(edStore, xid.StringId, protobuf.Validate)

		testSystemCreation(t, ctx, systemMgr)
	})
}

func testSystemCreation(t *testing.T, ctx context.Context, systemMgr *domain.SystemManager) *ponixv1.System {
	mockSystem := mockSystem(ponixv1.SystemStatus_SYSTEM_STATUS_PENDING)
	systemId, err := systemMgr.CreateSystem(ctx, mockSystem)
	if err != nil {
		t.Fatal(err)
	}

	system, err := systemMgr.System(ctx, systemId)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, mockSystem.Name, system.Name)
	assert.Equal(t, mockSystem.OrganizationId, system.OrganizationId)
	assert.Equal(t, mockSystem.Id, systemId)
	assert.Equal(t, mockSystem.Status, system.Status)

	return system
}

func setupDatabase(t *testing.T, ctx context.Context) (*tcp.PostgresContainer, *sqlc.Queries) {
	pnx := "ponix"

	postgresContainer, err := tcp.Run(ctx,
		"postgres:16-alpine",
		tcp.WithDatabase(pnx),
		tcp.WithUsername(pnx),
		tcp.WithPassword(pnx),
		tcp.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatal(err)
	}

	tcurl, err := tcUrl(ctx, postgresContainer)
	if err != nil {
		testcontainers.CleanupContainer(t, postgresContainer)
		t.Fatal(err)
	}

	curl := postgres.NewConnUrl(
		postgres.WithDB(pnx),
		postgres.WithUser(pnx),
		postgres.WithPassword(pnx),
		postgres.WithUrl(tcurl),
	)

	err = postgres.RunMigrations(ctx, curl)
	if err != nil {
		testcontainers.CleanupContainer(t, postgresContainer)
		t.Fatal(err)
	}

	pool, err := postgres.NewPool(ctx, curl)
	if err != nil {
		testcontainers.CleanupContainer(t, postgresContainer)
		t.Fatal(err)
	}

	return postgresContainer, sqlc.New(pool)
}

func tcUrl(ctx context.Context, pc *tcp.PostgresContainer) (string, error) {
	containerPort, err := pc.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return "", err
	}

	host, err := pc.Host(ctx)
	if err != nil {
		return "", err
	}

	return net.JoinHostPort(host, containerPort.Port()), nil
}

func mockSystem(status ponixv1.SystemStatus) *ponixv1.System {
	return ponixv1.System_builder{
		Id:             xid.StringId(),
		OrganizationId: xid.StringId(),
		Name:           gofakeit.Name(),
		Status:         status,
	}.Build()
}
