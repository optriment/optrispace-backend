package intest

import (
	"testing"

	"github.com/lithammer/shortuuid/v4"
	"github.com/stretchr/testify/assert"
)

func TestUuidGen(t *testing.T) {
	assert.Len(t, shortuuid.New(), 22)
}
