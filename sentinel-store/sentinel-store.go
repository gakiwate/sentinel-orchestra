package sentinelstore

import (
	"fmt"
	"log"
	"strconv"

	"github.com/cockroachdb/pebble"
)

type SentinelStore struct {
	DB *pebble.DB
}

func NewSentinelCounterStore(storeName string) *SentinelStore {
	var opts pebble.Options
	opts.Merger = &pebble.Merger{
		Merge: func(key, value []byte) (pebble.ValueMerger, error) {
			res := &CounterValueMerger{}
			val, _ := strconv.Atoi(string(value))
			res.Count += val
			return res, nil
		},
		Name: "SentinelCounterStore",
	}
	db, err := pebble.Open(fmt.Sprintf("%s.db", storeName), &opts)
	if err != nil {
		log.Fatal("error opening database:", err)
	}
	return &SentinelStore{
		DB: db,
	}
}
