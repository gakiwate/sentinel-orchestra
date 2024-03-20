package sentineldb

import (
	"fmt"

	"github.com/cockroachdb/pebble"
	sentinelstore "github.com/gakiwate/sentinel-orchestra/sentinel-store"
)

type SentinelDB struct {
	store     sentinelstore.SentinelStore
	StoreName string
}

// New Instance of SentinelDB
func NewTestSentinelDB(name string) *SentinelDB {

	return &SentinelDB{
		store:     *sentinelstore.NewSentinelDB(name, true),
		StoreName: name,
	}
}

// New Instance of SentinelDB
func NewSentinelDB(name string, tmpDB bool) *SentinelDB {

	return &SentinelDB{
		store:     *sentinelstore.NewSentinelDB(name, tmpDB),
		StoreName: name,
	}
}

// Close and release resources used by SentinelDB
func (db *SentinelDB) Close() error {
	return db.store.Close()
}

func (db *SentinelDB) AddResult(key string, resultJSON []byte) error {
	// Use a batch to perform the append operation atomically
	batch := db.store.DB.NewBatch()
	defer batch.Close()

	resultJSON = append(resultJSON, '\n')

	// Append the new JSON data as a line in the value associated with the key
	if err := batch.Merge([]byte(key), resultJSON, pebble.Sync); err != nil {
		return fmt.Errorf("error closing database: %w", err)
	}

	// Commit the batch to apply the changes
	return batch.Commit(pebble.Sync)
}

func (db *SentinelDB) Get(key string) ([]byte, error) {
	value, ioc, err := db.store.DB.Get([]byte(key))
	if err != nil {
		return []byte{}, nil
	}
	defer ioc.Close()
	return value, err
}
