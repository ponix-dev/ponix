package protobuf

import (
	"github.com/bufbuild/protovalidate-go"
	"github.com/ponix-dev/ponix/internal/domain"
	"google.golang.org/protobuf/proto"
)

// Validate checks that a protobuf message satisfies all protovalidate constraints.
// It returns domain.ErrInvalidMessageFormat if the input is not a valid proto.Message,
// or a validation error if the message fails any of its defined validation rules.
func Validate(msg any) error {
	pmsg, ok := msg.(proto.Message)
	if !ok {
		return domain.ErrInvalidMessageFormat
	}

	err := protovalidate.Validate(pmsg)
	if err != nil {
		return err
	}

	return nil
}
