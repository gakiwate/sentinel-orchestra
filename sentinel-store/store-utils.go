package sentinelstore

import (
	"io"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// CounterValueMerger is a merger that implements an associative merge operation for
// incrementing a counter.
type CounterValueMerger struct {
	Count int
}

// MergeNewer adds value to the result.
func (c *CounterValueMerger) MergeNewer(value []byte) error {
	val, err := strconv.Atoi(string(value))
	if err != nil {
		log.Error("merge newer:", err, "; value:", val)
	}
	c.Count += val
	return nil
}

// MergeOlder adds result to value.
func (c *CounterValueMerger) MergeOlder(value []byte) error {
	val, err := strconv.Atoi(string(value))
	if err != nil {
		log.Error("merge older:", err, "; value:", val)
	}
	c.Count += val
	return nil
}

func (c *CounterValueMerger) Finish(includesBase bool) ([]byte, io.Closer, error) {
	return []byte(strconv.Itoa(c.Count)), nil, nil
}
