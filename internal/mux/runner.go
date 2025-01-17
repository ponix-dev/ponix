package mux

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ponix-dev/ponix/internal/runner"
)

func NewRunner(srv *Server) runner.RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			errChan := make(chan error)

			go func() {
				srv.logger.Info("starting http server", slog.String("port", srv.port))
				err := srv.httpServer.ListenAndServe()
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					errChan <- err
				}

				defer close(errChan)
			}()

			select {
			case <-ctx.Done():
				return nil
			case err := <-errChan:
				return err
			}
		}
	}
}

func NewCloser(srv *Server) runner.RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			srv.logger.Info("stopping http server", slog.String("port", srv.port))
			err := srv.httpServer.Shutdown(ctx)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}

			return nil
		}
	}
}
