package sentinelutils

import (
	"testing"
)

func InitTest(t *testing.T) *SentinelCounters {
	return NewTestSentinelCounter("sentinel-utils-test")
}

func TestGetOnly(t *testing.T) {
	counters := InitTest(t)
	val, _ := counters.Get("cat1.test")
	if val != 0 {
		t.Errorf("Expected 0 but counter value is %d", val)
	}
}

func TestSimpleIncrement(t *testing.T) {
	counters := InitTest(t)
	counters.Incr("cat1.test")
	val, _ := counters.Get("cat1.test")
	if val != 1 {
		t.Errorf("Expected 1 but counter value is %d", val)
	}
}

func TestPrefixFuncs(t *testing.T) {
	counters := InitTest(t)
	counters.Incr("cat1.test")
	counters.Incr("cat2.test")
	counters.Incr("cat1.test")
	iter := counters.FetchAllKeysIterator([]byte("cat1"))
	for iter.First(); iter.Valid(); iter.Next() {
		if string(iter.Key()) != "cat1.test" || string(iter.Value()) != "2" {
			t.Errorf("Expected key cat1.test got %s. Expected value 2 got %s\n", string(iter.Key()), string(iter.Value()))
		}
	}
}

func TestPrefixData(t *testing.T) {
	counters := InitTest(t)
	counters.Incr("cat1.test")
	counters.Incr("cat2.test")
	counters.Incr("cat1.test")
	data := counters.FetchData([]byte("cat1"))
	if len(data) != 1 {
		t.Errorf("Expected length as 1 but got %d\n", len(data))
	}
	for k, v := range data {
		if k != "cat1.test" {
			t.Errorf("Expected cat1.test as key but got %s\n", k)
		}
		if v != 2 {
			t.Errorf("Expected 2 as value but got %d\n", v)
		}
	}
	data = counters.FetchData(nil)
	if len(data) != 2 {
		t.Errorf("Expected length as 2 but got %d\n", len(data))
	}
}
