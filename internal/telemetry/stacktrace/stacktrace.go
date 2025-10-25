package stacktrace

import (
	"fmt"
	"log/slog"
)

// NewStackTraceErrorf creates a new formatted error with captured stack trace information
// for enhanced observability in logs and traces.
func NewStackTraceErrorf(format string, a ...any) error {
	return new(fmt.Errorf(format, a...))
}

// NewStackTraceError wraps an existing error with captured stack trace information
// for enhanced observability in logs and traces.
func NewStackTraceError(err error) error {
	return new(err)
}

// ReplaceAttr is a slog handler function that enriches error attributes with stacktrace
// information when logging errors to improve observability.
func ReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	switch a.Value.Kind() {
	case slog.KindAny:
		switch v := a.Value.Any().(type) {
		case error:
			a.Value = fmtErr(v)
		}
	}

	return a
}

// FmtErr returns a slog.GroupValue with keys "msg" and "trace". If the error
// was not generated from NewStackTraceError, the "trace" key is omitted.
func fmtErr(err error) slog.Value {
	var groupValues []slog.Attr

	groupValues = append(groupValues, slog.String("msg", err.Error()))

	goErr, ok := err.(*Error)

	if ok {
		stack := goErr.StackTrace()

		groupValues = append(groupValues,
			slog.Any("stacktrace", stack),
		)
	}

	return slog.GroupValue(groupValues...)
}
