package zdnsorc

import (
	"encoding/binary"
	"sync"

	"github.com/cockroachdb/pebble"
	log "github.com/sirupsen/logrus"
)

type ZDNSMonitor struct {
	db     *pebble.DB
	dbLock sync.Mutex
}

// NewZDNSMonitor creates a new ZNDSMonitor.
func NewZDNSMonitor(dirname string) *ZDNSMonitor {
	db, err := pebble.Open(dirname, &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	return &ZDNSMonitor{
		db: db,
	}
}

// AddIP records that an IP address has been passed to ZGrab.
func (monitor *ZDNSMonitor) AddIP(addr string) {
	monitor.incrementValue(addr)
}

// CheckZDNSResult reviews a ZDNSResult and records errors accordingly.
func (monitor *ZDNSMonitor) CheckZDNSResult(result *ZDNSResult) {
	if result.Status != "NOERROR" {
		monitor.incrementValue("errors")
	}
	monitor.incrementValue("results")
}

// CloseMonitor cleans up ZDNSMonitor, closing the Pebble database.
func (monitor *ZDNSMonitor) CloseMonitor() {
	if err := monitor.db.Close(); err != nil {
		log.Fatal(err)
	}
}

func (monitor *ZDNSMonitor) incrementValue(key string) {
	monitor.dbLock.Lock()
	defer monitor.dbLock.Unlock()
	newValue := uint64(1)
	count, closer, err := monitor.db.Get([]byte(key))
	// Value found
	if err == nil {
		newValue += byteArrayToUint64(count)
		// Close to avoid memory leak
		if err := closer.Close(); err != nil {
			log.Fatal(err)
		}
	} else if err != pebble.ErrNotFound {
		log.Fatal(err)
	}

	err = monitor.db.Set([]byte(key), uint64ToByteArray(newValue), pebble.Sync)
	if err != nil {
		log.Fatal(err)
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
