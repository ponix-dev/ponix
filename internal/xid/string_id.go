package xid

import "github.com/rs/xid"

// StringId generates a globally unique string identifier using the xid library.
// The returned ID is sortable by time and URL-safe, making it suitable for use
// as database primary keys or distributed system identifiers.
func StringId() string {
	return xid.New().String()
}
