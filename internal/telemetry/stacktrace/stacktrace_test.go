package stacktrace_test

import (
	"testing"

	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

func TestReplaceAttr(t *testing.T) {
	t.Run("", func(t *testing.T) {
		stacktrace.ReplaceAttr([]string{}, stacktrace.ErrorAttribute(stacktrace.NewStackTraceErrorf("this is an error now")))
	})
}
