package sentinelmon

// Add more types as we track more fields
type ZDNSResult struct {
	Data     interface{}
	MetaData interface{}
	Status   string `json:"status"`
}
