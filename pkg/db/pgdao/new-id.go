package pgdao

import "github.com/lithammer/shortuuid/v4"

// NewID generates new unique ID
func NewID() string {
	return shortuuid.New()
}
