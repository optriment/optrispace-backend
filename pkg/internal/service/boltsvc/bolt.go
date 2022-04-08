package boltsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
)

var boltDB *bolt.DB

// openBolt opens bolt DB and saves it as a global variable.
// when context will be canceled, DB should be closed
func openBolt(ctx context.Context, path string) (*bolt.DB, error) {
	var err error
	if boltDB != nil {
		return boltDB, nil
	}

	boltDB, err = bolt.Open(path, 0o666, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("unable to open file for bolt DB: %w", err)
	}

	go func() {
		<-ctx.Done()
		log.Debug().Msg("Closing the boltDB")
		boltDB.Close()
		boltDB = nil
	}()
	return boltDB, nil
}

// generates new unique ID
func newID() string {
	return shortuuid.New()
}
