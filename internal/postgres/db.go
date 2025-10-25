package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a new PostgreSQL connection pool using the provided connection URL.
func NewPool(ctx context.Context, url ConnUrl) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(ctx, string(url))
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}

// ConnUrl represents a PostgreSQL connection string.
type ConnUrl string

type connUrlConfig struct {
	user     string
	password string
	url      string
	db       string
	tls      string
}

// ConnUrlOption is a functional option for configuring connection URL parameters.
type ConnUrlOption func(*connUrlConfig)

// WithUrl sets the host and port portion of the connection URL.
func WithUrl(url string) ConnUrlOption {
	return func(cuc *connUrlConfig) {
		cuc.url = url
	}
}

// WithUser sets the database user for the connection.
func WithUser(user string) ConnUrlOption {
	return func(cuc *connUrlConfig) {
		cuc.user = user
	}
}

// WithPassword sets the database password for the connection.
func WithPassword(password string) ConnUrlOption {
	return func(cuc *connUrlConfig) {
		cuc.password = password
	}
}

// WithDB sets the database name for the connection.
func WithDB(db string) ConnUrlOption {
	return func(cuc *connUrlConfig) {
		cuc.db = db
	}
}

var (
	tlsTrue  = "require"
	tlsFalse = "disable"
)

// WithTLS enables TLS/SSL for the database connection.
func WithTLS() ConnUrlOption {
	return func(cuc *connUrlConfig) {
		cuc.tls = tlsTrue
	}
}

//TODO: handle pool stuff

// NewConnUrl constructs a PostgreSQL connection URL from the provided options.
func NewConnUrl(options ...ConnUrlOption) ConnUrl {
	curlc := &connUrlConfig{
		user:     "postgres",
		password: "postgres",
		db:       "postgres",
		url:      "localhost:5432",
		tls:      tlsFalse,
	}

	for _, option := range options {
		option(curlc)
	}

	return ConnUrl(fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", curlc.user, curlc.password, curlc.url, curlc.db, curlc.tls))
}
