package sentinelmon

import (
	"encoding/json"
	"net/http"

	utils "github.com/gakiwate/sentinel-orchestra/sentinel-utils"
)

type SentinelMonitor struct {
	Stats utils.SentinelCounters
}

func (mon *SentinelMonitor) Serve() error {

	statsHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/stats" {
			// Set the response headers
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Send the JSON data as the response body
			data := mon.Stats.FetchData(nil)
			jsonData, _ := json.Marshal(data)
			w.Write(jsonData)
		} else {
			// If the URL is not recognized, return a 404 error
			http.NotFound(w, r)
		}
	}

	http.HandleFunc("/", statsHandler)
	return http.ListenAndServe(":8000", nil)
}

func NewSentinelMonitor() *SentinelMonitor {
	return &SentinelMonitor{
		Stats: *utils.NewSentinelCounter("sentinel-stats", false),
	}
}
