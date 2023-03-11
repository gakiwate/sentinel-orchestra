package sentinelutils

import (
	"strconv"

	"github.com/cockroachdb/pebble"
	sentinelstore "github.com/gakiwate/sentinel-orchestra/sentinel-store"
	log "github.com/sirupsen/logrus"
)

type SentinelCounters struct {
	store *sentinelstore.SentinelStore
}

func NewTestSentinelCounter(storeName string) *SentinelCounters {
	return NewSentinelCounter(storeName, true)
}

func NewSentinelCounter(storeName string, tmpDB bool) *SentinelCounters {
	return &SentinelCounters{
		store: sentinelstore.NewSentinelCounterStore(storeName, tmpDB),
	}
}

func (ctrdb *SentinelCounters) Incr(key string) error {
	ctrdb.store.DB.Merge([]byte(key), []byte("1"), pebble.NoSync)
	return nil
}

func (ctrdb *SentinelCounters) Get(key string) (int, error) {
	value, ioc, err := ctrdb.store.DB.Get([]byte(key))
	if err != nil {
		log.Info("creating new value for key:", key)
		return 0, ctrdb.store.DB.Set([]byte(key), []byte("0"), pebble.Sync)
	}
	defer ioc.Close()
	val, err := strconv.Atoi(string(value))
	return val, err
}

func (ctrdb *SentinelCounters) FetchData(keyPrefix []byte) map[string]int {
	data := make(map[string]int)
	iter := ctrdb.FetchAllKeysIterator(keyPrefix)
	for iter.First(); iter.Valid(); iter.Next() {
		k := string(iter.Key())
		val, _ := strconv.Atoi(string(iter.Value()))
		data[k] = val
	}
	return data
}

func (ctrdb *SentinelCounters) FetchAllKeysIterator(keyPrefix []byte) *pebble.Iterator {
	keyUpperBound := func(b []byte) []byte {
		end := make([]byte, len(b))
		copy(end, b)
		for i := len(end) - 1; i >= 0; i-- {
			end[i] = end[i] + 1
			if end[i] != 0 {
				return end[:i+1]
			}
		}
		return nil // no upper-bound
	}

	prefixIterOptions := func(prefix []byte) *pebble.IterOptions {
		return &pebble.IterOptions{
			LowerBound: keyPrefix,
			UpperBound: keyUpperBound(keyPrefix),
		}
	}
	return ctrdb.store.DB.NewIter(prefixIterOptions(keyPrefix))
}
