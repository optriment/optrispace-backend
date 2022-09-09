package pgdao

import (
	"context"

	"github.com/lithammer/shortuuid/v4"
)

// NewID generates new unique ID
func NewID() string {
	return shortuuid.New()
}

// PurgeDB purges ALL data from the DB.
// In testing purposes only!
// HANDLE WITH CARE!
func PurgeDB(ctx context.Context, db DBTX) error {
	queries := New(db)

	if e := queries.MessagesPurge(ctx); e != nil {
		return e
	}

	if e := queries.ChatsPurge(ctx); e != nil {
		return e
	}

	if e := queries.ContractsPurge(ctx); e != nil {
		return e
	}

	if e := queries.ApplicationsPurge(ctx); e != nil {
		return e
	}

	if e := queries.JobsPurge(ctx); e != nil {
		return e
	}

	if e := queries.PersonsPurge(ctx); e != nil {
		return e
	}

	return nil
}
