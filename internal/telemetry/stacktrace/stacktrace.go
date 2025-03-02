package stacktrace

import (
	"fmt"
	"log/slog"
)

// NewStackTraceErr starts a stack trace with the given inputs.  It can either be used
// to generate a new custom error.
func NewStackTraceErrorf(format string, a ...any) error {
	return new(fmt.Errorf(format, a...))
}

// NewStackTraceErr starts a stack trace with the given inputs.  It can be used
// to wrap an existing error.
func NewStackTraceError(err error) error {
	return new(err)
}

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
