package xid

import "github.com/rs/xid"

func StringId() string {
	return xid.New().String()
}
