package stacktrace

import (
	"fmt"
	"log/slog"
	"runtime"
	"sync"
)

// MaxStackDepth defines the maximum number of stack frames captured when creating
// a stacktrace error for observability purposes.
const MaxStackDepth = 50

// StackSkipCount defines the number of stack frames to skip from the beginning of
// the stack to exclude runtime and stacktrace package internal calls.
const StackSkipCount = 3

// Error wraps an error with captured stack frame information for enhanced error
// observability and debugging in logs and traces.
type Error struct {
	Err          error
	frames       *runtime.Frames
	cachedFrames []string
	cached       bool
	mutex        *sync.Mutex
}

// new creates a new stacktrace error that contains the original error as well as set of call
func new(e error) error {
	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(3, stack[:])

	return &Error{
		Err:    e,
		frames: runtime.CallersFrames(stack[:length]),
		mutex:  &sync.Mutex{},
	}
}

// Error implements error interface by returning the orginal errors value
func (err *Error) Error() string {
	return err.Err.Error()
}

// Unwrap returns the underlying error that triggered a stack trace.  This is helpful for methods like errors.Is and errors.As
func (err *Error) Unwrap() error {
	return err.Err
}

// StackTrace returns a slice of formattted strings that contains each call in the stack's function, file, and file line.  If StackTrace has already been called,
// the stacktrace of the error is cached to support multiple callers.
func (err *Error) StackTrace() []string {
	err.mutex.Lock()
	defer err.mutex.Unlock()
	st := []string{}

	if err.cached {
		return err.cachedFrames
	}

	for {
		frame, more := err.frames.Next()

		st = append(st, fmt.Sprintf("%s: %s:%d\n", frame.Function, frame.File, frame.Line))

		if !more {
			break
		}
	}

	err.cachedFrames = st
	err.cached = true

	return st
}

// ErrorAttribute creates a structured log attribute for the given error, suitable
// for inclusion in OpenTelemetry log records.
func ErrorAttribute(err error) slog.Attr {
	return slog.Any("err", err)
}
