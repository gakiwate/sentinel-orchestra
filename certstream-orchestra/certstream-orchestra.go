package certstreamorc

import (
	"fmt"
	"strings"
	"time"

	"github.com/CaliDog/certstream-go"
	mon "github.com/gakiwate/sentinel-orchestra/sentinel-monitor"
	"github.com/nsqio/go-nsq"

	log "github.com/sirupsen/logrus"
)

// SentinelCertstreamOrchestrator is the orchestrator for the certstream program
type SentinelCertstreamOrchestrator struct {
	monitor     *mon.SentinelMonitor
	nsqHost     string
	nsqOutTopic string
}

// NewSentinelCertstreamOrchestrator creates a new SentinelCertstreamOrchestrator
func NewSentinelCertstreamOrchestrator(monitor *mon.SentinelMonitor, nsqHost string, nsqOutTopic string) *SentinelCertstreamOrchestrator {
	return &SentinelCertstreamOrchestrator{
		monitor:     monitor,
		nsqHost:     nsqHost,
		nsqOutTopic: nsqOutTopic,
	}
}

func formatFQDN(s string) string {
	// Remove wild card prefixes
	return strings.TrimPrefix(s, "*.")
}

func formatSHA1(s string) string {
	// remove : from the string and make lower case
	s = strings.ReplaceAll(s, ":", "")
	return strings.ToLower(s)
}

func removeDuplicates(strArray []string) []string {
	strMap := make(map[string]bool)
	for _, s := range strArray {
		strMap[s] = true
	}

	set := []string{}
	for s := range strMap {
		set = append(set, s)
	}

	return set
}

// Run starts the SentinelCertstreamOrchestrator
func (o *SentinelCertstreamOrchestrator) Run() {

	var nsqHost string = o.nsqHost
	var nsqOutTopic string = o.nsqOutTopic

	// Set Logger Level
	log.SetLevel(log.ErrorLevel)

	// Create a new NSQ producer

	nsqUrl := fmt.Sprintf("%s:4150", nsqHost)
	producer, err := nsq.NewProducer(nsqUrl, nsq.NewConfig())
	producer.SetLoggerLevel(nsq.LogLevelError)
	if err != nil {
		// Report Error and Exit.
		log.Fatal(err)
	}
	log.Info(fmt.Sprintf("Connecting to NSQ at %s", nsqUrl))

	stream, errStream := certstream.CertStreamEventStream(false)
	for {
		o.monitor.Stats.Incr("certstream.cert_cnt")
		select {
		case jq := <-stream:
			domains, err := jq.ArrayOfStrings("data", "leaf_cert", "all_domains")
			// format all domains to remove wildcard entries and lower case
			for idx, domain := range domains {
				domains[idx] = formatFQDN(domain)
			}
			// remove duplicates. primarily as a result of wild card removals
			domains = removeDuplicates(domains)
			if err != nil {
				log.Error(err)
			}

			// get cert sha1
			certSHA1, err := jq.String("data", "leaf_cert", "fingerprint")
			if err != nil {
				log.Error(err)
			}
			certSHA1 = formatSHA1(certSHA1)

			for _, domain := range domains {
				o.monitor.Stats.Incr("certstream.domain_cnt")
				tnow := time.Now().Unix()
				zdnsFeedInput := fmt.Sprintf("{\"domain\": \"%s\",\"metadata\": {\"cert_sha1\": \"%s\", \"scan_after\": \"%d\"}}", domain, certSHA1, tnow)
				err = producer.Publish(nsqOutTopic, []byte(zdnsFeedInput))
				log.Info(fmt.Sprintf("Certstream: Publishing %s to channel %s", zdnsFeedInput, nsqOutTopic))
				if err != nil {
					log.Error(err)
				}
			}

		case err := <-errStream:
			o.monitor.Stats.Incr("certstream.cert_err_cnt")
			log.Error(err)
		}
	}
}
