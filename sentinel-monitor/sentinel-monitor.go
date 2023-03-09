package sentinelmon

import (
	utils "github.com/gakiwate/sentinel-orchestra/sentinel-utils"
)

type SentinelMonitor struct {
	Stats utils.SentinelCounters
}

func (mon *SentinelMonitor) MonitorLoop() error {
	return mon.Stats.FetchAll()
}

func NewSentinelMonitor() *SentinelMonitor {
	return &SentinelMonitor{
		Stats: *utils.NewSentinelCounter("sentinel-stats"),
	}
}
