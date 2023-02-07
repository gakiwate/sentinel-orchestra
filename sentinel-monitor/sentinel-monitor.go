package sentinelmon

import (
	"encoding/binary"
	"sync"

	"github.com/cockroachdb/pebble"
	log "github.com/sirupsen/logrus"
)

type SentinelMonitor struct {
	dbLock sync.Mutex
	db     *pebble.DB
}

// NewSentinelMonitor creates a new SentinelMonitor.
func NewSentinelMonitor(dirname string) *SentinelMonitor {
	db, err := pebble.Open(dirname, &pebble.Options{})
	if err != nil {
		log.Error(err)
	}
	return &SentinelMonitor{
		db: db,
	}
}

// AddIP records that an IP address has been passed to ZGrab.
func (monitor *SentinelMonitor) AddIP(addr string) {
	monitor.incrementValue(addr)
}

// CheckZDNSResult reviews a ZDNSResult and records errors accordingly.
func (monitor *SentinelMonitor) CheckZDNSResult(result *ZDNSResult) {
	if result.Status != "NOERROR" {
		monitor.incrementValue("errors")
	}
	monitor.incrementValue("results")
}

// CloseMonitor cleans up SentinelMonitor, closing the Pebble database.
func (monitor *SentinelMonitor) CloseMonitor() {
	if err := monitor.db.Close(); err != nil {
		log.Fatal(err)
	}
}

func (monitor *SentinelMonitor) incrementValue(key string) {
	monitor.dbLock.Lock()
	defer monitor.dbLock.Unlock()
	newValue := uint64(1)
	count, closer, err := monitor.db.Get([]byte(key))
	// Value found
	if err == nil {
		newValue += byteArrayToUint64(count)
		// Close to avoid memory leak
		if err := closer.Close(); err != nil {
			log.Error(err)
		}
	} else if err != pebble.ErrNotFound {
		log.Error(err)
	}

	err = monitor.db.Set([]byte(key), uint64ToByteArray(newValue), pebble.Sync)
	if err != nil {
		log.Error(err)
	}
}

func uint64ToByteArray(i uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return b
}

func byteArrayToUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}
