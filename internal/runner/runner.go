package runner

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"golang.org/x/sync/errgroup"
)

// RunnerFunc is a function that takes in a context and returns a function that is configured to return when an error occurs or the passed in Context is finished.
type RunnerFunc func(ctx context.Context) func() error

type RunnerOption func(*Runner)

// WithAppProcess adds a RunnerFunc that will run until it returns an error or the application receives a SIGTERM signal from the host.
// It is up to the developer to make sure that their RunnerFunc handles context properly, otherwise, their application has the potential to hang while the runner waits for the process to complete.
func WithAppProcess(rf RunnerFunc) RunnerOption {
	return func(r *Runner) {
		r.appProcesses = append(r.appProcesses, rf)
	}
}

// WithCloser adds a RunnerFunc that will run after all app processes have returned.  All closers will be attempted regardless of app process error.
func WithCloser(rf RunnerFunc) RunnerOption {
	return func(r *Runner) {
		r.closers = append(r.closers, rf)
	}
}

// WithCloserTimeout configures how long the runner will wait for all closers to complete before timing out.  The default timeout is 10 seconds.
func WithCloserTimeout(t time.Duration) RunnerOption {
	return func(r *Runner) {
		r.closeTimeout = t
	}
}

// WithLogger configures the logger that will be used to report any errors from app processes and closers.  The default logger will write any errors to stdout.
func WithLogger(l *slog.Logger) RunnerOption {
	return func(r *Runner) {
		r.logger = l
	}
}

func WithContext(ctx context.Context) RunnerOption {
	return func(r *Runner) {
		r.ctx = ctx
	}
}

// Runner handles long running app processes and closing functions when stopping an application
type Runner struct {
	appProcesses []RunnerFunc
	closers      []RunnerFunc
	closeTimeout time.Duration
	logger       *slog.Logger
	ctx          context.Context
}

// NewRunner takes in RunnerOptions and configures a Runner for application processes.  App processes and closers will be called in the order they are passed in.
func New(options ...RunnerOption) *Runner {
	r := &Runner{
		appProcesses: make([]RunnerFunc, 0),
		closers:      make([]RunnerFunc, 0),
		closeTimeout: time.Second * 10,
		logger:       slog.New(slog.NewTextHandler(os.Stdout, nil)),
		ctx:          context.Background(),
	}

	for _, o := range options {
		o(r)
	}

	return r
}

// Run will call its app processes so that they run concurrently until one of the processes returns an error or the host sends the app a SIGTERM signaling shutdown.
// Run will always attempt its closers regardless of app shutdown being triggered due to an error or SIGTERM signal.
func (r *Runner) Run() {
	rg, rgCtx := newSignalErrGroup(r.ctx, r.logger)

	r.logger.Info("starting app runners...")

	for _, ap := range r.appProcesses {
		rg.Go(ap(rgCtx))
	}

	err := rg.Wait()
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			r.logger.Error("error stopping runners", slog.String("err", err.Error()))
		}
	}

	cg, cgCtx := newTimeoutErrGroup(r.ctx, r.closeTimeout)

	for _, closer := range r.closers {
		cg.Go(closer(cgCtx))
	}

	err = cg.Wait()
	if err != nil {
		r.logger.Error("error stopping closers", slog.String("err", err.Error()))
	}

	r.logger.Info("good bye!")
	os.Exit(0)
}

// newSignalErrGroup creates an errgroup that is configured to cancel its context and shutdown when a SIGTERM signal is sent from
// the host to your application.  This errgroup will run until SIGTERM is sent or a process exits with an error.
func newSignalErrGroup(ctx context.Context, logger *slog.Logger) (*errgroup.Group, context.Context) {
	cancelCtx, cancel := context.WithCancel(ctx)
	eg, egCtx := errgroup.WithContext(cancelCtx)
	eg.Go(gracefulShutdownRunner(egCtx, cancel, logger))
	return eg, egCtx
}

// gracefulShutdownRunner implements RunFunc where the passed in cancel func is called when SIGTERM is sent from the host to the application.
func gracefulShutdownRunner(ctx context.Context, done context.CancelFunc, logger *slog.Logger) func() error {
	return func() error {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		select {
		case <-signalChannel:
			logger.Info("starting graceful shutdown...")
			done()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// newTimeoutErrGroup returns an errgroup configured to stop subprocesses after a given period time if they have not already completed.  If a process completes with an error or all processes complete without one,
// then the returned errgroup will complete normally.
func newTimeoutErrGroup(ctx context.Context, timeout time.Duration) (*errgroup.Group, context.Context) {
	closerCtx, _ := context.WithTimeout(ctx, timeout)
	return errgroup.WithContext(closerCtx)
}
