package mux

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type HttpServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

type ServeMux interface {
	Handle(pattern string, handler http.Handler)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

var (
	ErrNoServerPort = errors.New("http server address needs a port")
)

type Server struct {
	port       string
	logger     *slog.Logger
	httpServer HttpServer
}

type ServerOption func(config *serverOptionConfig)

type serverOptionConfig struct {
	mux         ServeMux
	http2Server *http2.Server
	port        string
	addr        string
	logger      *slog.Logger
}

func WithPort(port string) ServerOption {
	return func(config *serverOptionConfig) {
		config.port = port
		config.addr = fmt.Sprintf("0.0.0.0:%s", port)
	}
}

func WithLogger(lgr *slog.Logger) ServerOption {
	return func(srv *serverOptionConfig) {
		srv.logger = lgr
	}
}

func WithHandler(path string, handler http.Handler) ServerOption {
	return func(srv *serverOptionConfig) {
		srv.logger.Info("registering handler", slog.String("path", path))
		srv.mux.Handle(path, handler)
	}
}

// WithHttp2 specifies for the mux to be wrapped with golang.org/x/net/http2/h2c handler and adds on the default http2 server.
func WithHttp2() ServerOption {
	return func(srv *serverOptionConfig) {
		srv.http2Server = &http2.Server{}
	}
}

// WithHttp2Server specifies for the mux to be wrapped with golang.org/x/net/http2/h2c handler and adds on the provided http2 server.
func WithHttp2Server(server *http2.Server) ServerOption {
	return func(srv *serverOptionConfig) {
		srv.http2Server = server
	}
}

func New(mux ServeMux, options ...ServerOption) (*Server, error) {
	if mux == nil {
		mux = http.NewServeMux()
	}

	config := &serverOptionConfig{
		mux:    mux,
		logger: slog.Default(),
	}

	for _, o := range options {
		o(config)
	}

	if config.addr == "" && config.port == "" {
		return nil, ErrNoServerPort
	}

	var handler http.Handler
	if config.http2Server != nil {
		handler = h2c.NewHandler(config.mux, config.http2Server)
	} else {
		handler = config.mux
	}

	httpSrv := &http.Server{
		Handler: handler,
		Addr:    config.addr,
	}

	return &Server{
		port:       config.port,
		logger:     config.logger,
		httpServer: httpSrv,
	}, nil
}
