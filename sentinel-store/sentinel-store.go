package sentinelstore

import (
	"fmt"
	"log"
	"strconv"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
)

type SentinelStore struct {
	DB *pebble.DB
}

func NewSentinelCounterStore(storeName string, tmpDB bool) *SentinelStore {
	var opts pebble.Options
	opts.Merger = &pebble.Merger{
		Merge: func(key, value []byte) (pebble.ValueMerger, error) {
			res := &CounterValueMerger{}
			val, err := strconv.Atoi(string(value))
			if err != nil {
				return res, err
			}
			res.Count += val
			return res, nil
		},
		Name: "SentinelCounterStore",
	}
	if tmpDB {
		opts.FS = vfs.NewMem()
	}

	db, err := pebble.Open(fmt.Sprintf("%s.db", storeName), &opts)
	if err != nil {
		log.Fatal("error opening database:", err)
	}
	return &SentinelStore{
		DB: db,
	}
}

func (store *SentinelStore) Close() {
	store.DB.Close()
}
