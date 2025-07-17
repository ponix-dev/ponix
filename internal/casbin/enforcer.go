package casbin

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxadapter "github.com/pckhoi/casbin-pgx-adapter/v3"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

// Enforcer wraps Casbin enforcer for organization-based access control
type Enforcer struct {
	casbin *casbin.Enforcer
}

// NewEnforcer creates a new Casbin enforcer with the organization model and pgx adapter
func NewEnforcer(ctx context.Context, pool *pgxpool.Pool) (*Enforcer, error) {
	m := initializeModel()

	connStr := pool.Config().ConnString()
	a, err := pgxadapter.NewAdapter(connStr, pgxadapter.WithTableName("casbin_rule"))
	if err != nil {
		return nil, stacktrace.NewStackTraceErrorf("failed to create pgx adapter: %w", err)
	}

	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, stacktrace.NewStackTraceErrorf("failed to create enforcer: %w", err)
	}

	enforcer := &Enforcer{casbin: e}

	// Initialize default policies in code
	err = enforcer.initializePolicies(ctx)
	if err != nil {
		return nil, stacktrace.NewStackTraceErrorf("failed to initialize policies: %w", err)
	}

	return enforcer, nil
}

func initializeModel() model.Model {
	// Create model from code instead of file
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act, org")
	m.AddDef("p", "p", "sub, obj, act, org")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act && r.org == p.org")

	return m
}

// initializePolicies sets up the default role-based policies in code
func (e *Enforcer) initializePolicies(_ctx context.Context) error {
	// Clear existing policies to start fresh
	e.casbin.ClearPolicy()

	// Define base role policies for end devices and organizations
	policies := [][]string{
		// Admin role policies - full access across all organizations
		{"org_admin", "end_device", "create", "*"},
		{"org_admin", "end_device", "read", "*"},
		{"org_admin", "end_device", "update", "*"},
		{"org_admin", "end_device", "delete", "*"},
		{"org_admin", "organization", "read", "*"},
		{"org_admin", "organization", "update", "*"},
		{"org_admin", "user", "create", "*"},
		{"org_admin", "user", "update", "*"},
		{"org_admin", "user", "delete", "*"},
		{"org_admin", "user", "read", "*"},

		// Member role policies - read and update access
		{"org_member", "end_device", "read", "*"},
		{"org_member", "end_device", "update", "*"},
		{"org_member", "organization", "read", "*"},
		{"org_member", "user", "read", "*"},

		// Viewer role policies - read-only access
		{"org_viewer", "end_device", "read", "*"},
		{"org_viewer", "organization", "read", "*"},
		{"org_viewer", "user", "read", "*"},
	}

	// Add all policies
	for _, policy := range policies {
		_, err := e.casbin.AddPolicy(policy)
		if err != nil {
			return stacktrace.NewStackTraceErrorf("failed to add policy %v: %w", policy, err)
		}
	}

	// Save policies to database
	err := e.casbin.SavePolicy()
	if err != nil {
		return stacktrace.NewStackTraceErrorf("failed to save policies: %w", err)
	}

	return nil
}
