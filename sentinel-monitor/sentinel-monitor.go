package sentinelmon

import (
	"encoding/json"
	"net/http"

	matcher "github.com/gakiwate/sentinel-orchestra/sentinel-matcher"
	utils "github.com/gakiwate/sentinel-orchestra/sentinel-utils"
)

type SentinelMonitor struct {
	Matcher *matcher.Matcher
	Stats   utils.SentinelCounters
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

func NewSentinelMonitor(monitorName string, ipPrefixes []string, domains []string) *SentinelMonitor {
	return &SentinelMonitor{
		Matcher: matcher.NewMatcher(ipPrefixes, domains),
		Stats:   *utils.NewSentinelCounter(monitorName, false),
	}
}
