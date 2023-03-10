package sentinelmon

import (
	"fmt"
	"time"

	utils "github.com/gakiwate/sentinel-orchestra/sentinel-utils"
)

type SentinelMonitor struct {
	Stats utils.SentinelCounters
}

func (mon *SentinelMonitor) MonitorLoop() error {
	for {
		data := mon.Stats.FetchData(nil)
		for k, v := range data {
			// TODO: Replace with something that is not simple print
			fmt.Println(k, v)
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func NewSentinelMonitor() *SentinelMonitor {
	return &SentinelMonitor{
		Stats: *utils.NewSentinelCounter("sentinel-stats", false),
	}
}
