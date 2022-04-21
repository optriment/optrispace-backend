package intest

import (
	"testing"

	"github.com/lithammer/shortuuid/v4"
	"github.com/stretchr/testify/assert"
	"optrispace.com/work/pkg/service/pgsvc"
)

func TestUuidGen(t *testing.T) {
	assert.Len(t, shortuuid.New(), 22)
}

func TestPassword(t *testing.T) {
	pass := "1234asdf"
	hash := pgsvc.CreateHashFromPassword(pass)

	err := pgsvc.CompareHashAndPassword(hash, pass)
	assert.NoError(t, err)
}
