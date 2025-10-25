package stacktrace

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// otelHandlerWrapper wraps an OpenTelemetry slog handler to enhance error logging
// with stacktrace attributes for improved observability.
type otelHandlerWrapper struct {
	*otelslog.Handler
}

// NewOtelHandlerWrapper wraps an existing otelslog Handler with the ability to append a stacktrace for stacktrace supported Errors.
func NewOtelHandlerWrapper(handler *otelslog.Handler) *otelHandlerWrapper {
	return &otelHandlerWrapper{
		Handler: handler,
	}
}

func (ohw *otelHandlerWrapper) Handle(ctx context.Context, record slog.Record) error {
	var stacktraceAttr slog.Attr
	addStackTrace := false
	record.Attrs(func(a slog.Attr) bool {
		switch a.Value.Kind() {
		case slog.KindAny:
			switch v := a.Value.Any().(type) {
			case error:
				stacktraceAttr, addStackTrace = ErrStackTraceAttr(v)
				return false
			}
		}

		return true
	})

	if addStackTrace {
		record.AddAttrs(stacktraceAttr)
	}

	return ohw.Handler.Handle(ctx, record)
}

// ErrStackTraceAttr returns a stacktrace attribute if the given error supports it.  Helpful for slog handlers that do not support ReplaceAttr functionality.
func ErrStackTraceAttr(err error) (slog.Attr, bool) {
	goErr, ok := err.(*Error)

	if ok {
		stack := goErr.StackTrace()

		return slog.Any("stacktrace", stack), true
	}

	return slog.Attr{}, false
}
