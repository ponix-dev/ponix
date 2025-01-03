package protobuf

import (
	"github.com/bufbuild/protovalidate-go"
	"github.com/ponix-dev/ponix/internal/domain"
	"google.golang.org/protobuf/proto"
)

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
