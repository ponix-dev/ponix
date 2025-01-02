package runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestRunner(t *testing.T) {
	t.Run("returns when app process runs in to an error", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

		r := New(
			WithAppProcess(sleepRunner(logger, 1)),
			WithAppProcess(sleepRunner(logger, 2)),
			WithAppProcess(errorRunner()),
			WithCloser(sleepCloser(logger)),
			WithCloserTimeout(time.Second*5),
			WithLogger(logger),
		)

		r.Run()
	})

	t.Run("logs an error when closers return an error", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

		r := New(
			WithAppProcess(sleepRunner(logger, 1)),
			WithAppProcess(sleepRunner(logger, 2)),
			WithAppProcess(errorRunner()),
			WithCloser(sleepCloser(logger)),
			WithCloser(errorRunner()),
			WithCloserTimeout(time.Second*5),
			WithLogger(logger),
			WithContext(context.TODO()),
		)

		r.Run()
	})
}

func sleepRunner(logger *slog.Logger, runner int) RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					fmt.Println(runner)
					time.Sleep(time.Millisecond * 100)
				}
			}
		}
	}
}

func errorRunner() RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			return errors.New("there was an error")
		}
	}
}

func sleepCloser(logger *slog.Logger) RunnerFunc {
	return func(ctx context.Context) func() error {
		return func() error {
			logger.Info("closer sleeping for 3 seconds")
			time.Sleep(time.Millisecond * 100)
			return nil
		}
	}
}
