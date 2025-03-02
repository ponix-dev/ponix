package stacktrace

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	ErrTest = errors.New("test error")
)

func TestError_Unwrap(t *testing.T) {

	t.Run("returns base error when unwrapped", func(t *testing.T) {
		assert := assert.New(t)

		err := anotherOne()

		assert.ErrorIs(err, ErrTest)
	})
}

func anotherOne() error {
	err := someFunc()
	if err != nil {
		return err
	}

	return nil
}

func someMethod1() error {
	return nil
}

var (
	ErrOops = errors.New("oops")
)

func someMethod2() error {
	return NewStackTraceError(ErrTest)
}

func someMethod3() error {
	return nil
}

func someFunc() error {
	err := someMethod1()
	if err != nil {
		return err
	}

	err = someMethod2()
	if err != nil {
		return err
	}

	err = someMethod3()
	if err != nil {
		return err
	}

	return nil
}
