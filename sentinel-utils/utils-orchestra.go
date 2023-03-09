package sentinelutils

import (
	"fmt"
	"strconv"

	"github.com/cockroachdb/pebble"
	sentinelstore "github.com/gakiwate/sentinel-orchestra/sentinel-store"
)

type SentinelCounters struct {
	store *sentinelstore.SentinelStore
}

func NewSentinelCounter(storeName string) *SentinelCounters {
	return &SentinelCounters{
		store: sentinelstore.NewSentinelCounterStore(storeName),
	}
}

func (ctrdb *SentinelCounters) Incr(key string) error {
	ctrdb.store.DB.Merge([]byte(key), []byte("1"), pebble.NoSync)
	return nil
}

func (ctrdb *SentinelCounters) Get(key string) (int, error) {
	value, ioc, err := ctrdb.store.DB.Get([]byte(key))
	defer ioc.Close()
	if err != nil {
		fmt.Println("creating new value for key:", key)
		ctrdb.store.DB.Set([]byte(key), []byte("0"), pebble.Sync)
		return 0, nil
	}
	val, err := strconv.Atoi(string(value))
	return val, err
}

func (ctrdb *SentinelCounters) FetchAll() error {
	return nil
}
