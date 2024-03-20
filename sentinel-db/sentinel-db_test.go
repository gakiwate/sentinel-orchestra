package sentineldb

import (
	"bytes"
	"testing"
)

func InitTest(t *testing.T) *SentinelDB {
	return NewTestSentinelDB("sentinel-db-test")
}

func TestEmptyFQDN(t *testing.T) {
	db := InitTest(t)
	data, _ := db.Get("www.invalid.domain")

	if len(data) != 0 {
		t.Errorf("Expected value to be empty but is %s", data)
	}
}

func TestAddResult1(t *testing.T) {
	db := InitTest(t)
	value := []byte(`{"name": "www.valid.domain", "add": "1"}`)
	key := "www.valid.domain"
	err := db.AddResult(key, value)
	if err != nil {
		t.Error("Unable to add to key")
	}

	data, _ := db.Get(key)
	// Add Result adds new line
	value = append(value, '\n')
	if !bytes.Equal(data, value) {
		t.Errorf("Expected %s but instead got %s", value, data)
	}

	value2 := []byte(`{"name": "www.valid.domain", "add": "2"}`)
	err = db.AddResult(key, value2)
	if err != nil {
		t.Error("Unable to add to key")
	}

	data, _ = db.Get(key)
	// Add Result adds new line
	value2 = append(value2, '\n')
	// Expected the concatentation of two values
	if !bytes.Equal(data, append(value, value2...)) {
		t.Errorf("Expected %s but instead got %s", value, data)
	}

}
