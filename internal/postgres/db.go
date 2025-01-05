package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, url ConnUrl) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(ctx, string(url))
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}

type ConnUrl string

type connUrlConfig struct {
	user     string
	password string
	url      string
	db       string
	tls      string
}

type ConnUrlOption func(*connUrlConfig)

func WithUrl(url string) ConnUrlOption {
	return func(cuc *connUrlConfig) {
		cuc.url = url
	}
}

func WithUser(user string) ConnUrlOption {
	return func(cuc *connUrlConfig) {
		cuc.user = user
	}
}

func WithPassword(password string) ConnUrlOption {
	return func(cuc *connUrlConfig) {
		cuc.password = password
	}
}

func WithDB(db string) ConnUrlOption {
	return func(cuc *connUrlConfig) {
		cuc.db = db
	}
}

var (
	tlsTrue  = "require"
	tlsFalse = "disable"
)

func WithTLS() ConnUrlOption {
	return func(cuc *connUrlConfig) {
		cuc.tls = tlsTrue
	}
}

//TODO: handle pool stuff

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
