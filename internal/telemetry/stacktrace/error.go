package stacktrace

import (
	"fmt"
	"log/slog"
	"runtime"
	"sync"
)

// The maximum number of stackframes on any error.
const MaxStackDepth = 50

// The amount of methods to remove from the beginning of the stack.  This helps remove calls to runtime package as well as calls in to the stacktrace package from the stack
const StackSkipCount = 3

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

func ErrorAttribute(err error) slog.Attr {
	return slog.Any("err", err)
}
