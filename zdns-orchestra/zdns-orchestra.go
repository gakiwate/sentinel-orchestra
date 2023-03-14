package zdnsorc

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"
)

type SentinelZDNSOrchestrator struct {
	nsqHost          string
	ipv4		 bool
	ipv6		 bool
	consumer         nsq.Consumer
	producer         nsq.Producer
	nsqZDNSOutTopic  string
	nsqZGrabOutTopic string
	zdnsDelay        int64
}

type ZDNSMetadata struct {
	CertSHA1  string `json:"cert_sha1"`
	ScanAfter string `json:"scan_after"`
}

type ZDNSResultData struct {
	BaseName      string   `json:"base_name"`
	Name          string   `json:"name"`
	IPv4Addresses []string `json:"ipv4_addresses"`
	IPv6Addresses []string `json:"ipv6_addresses"`
}

type ZDNSResult struct {
	Data     ZDNSResultData `json:"data"`
	MetaData ZDNSMetadata   `json:"metadata"`
}

type SentinelOrchestratorConfig struct {
	nsqHost          string
	ipv4		 bool
	ipv6		 bool
	nsqInTopic       string
	nsqZDNSOutTopic  string
	nsqZGrabOutTopic string
	zdnsDelay        int64
}

func NewSentinelZDNS4hrDelayOrchestrator(nsqHost string, ipv4 bool, ipv6 bool) *SentinelZDNSOrchestrator {
	cfg4hr := &SentinelOrchestratorConfig{
		nsqHost:          nsqHost,
		ipv4:		  ipv4,
		ipv6:		  ipv6,
		nsqInTopic:       "zdns_results",
		nsqZDNSOutTopic:  "zdns_4hr",
		nsqZGrabOutTopic: "zgrab",
		zdnsDelay:        14400, // 4hours -- 3600 sec * 4
	}
	return NewSentinelZDNSOrchestrator(*cfg4hr)
}

func NewSentinelZDNS24hrDelayOrchestrator(nsqHost string, ipv4 bool, ipv6 bool) *SentinelZDNSOrchestrator {
	cfg24hr := &SentinelOrchestratorConfig{
		nsqHost:          nsqHost,
		ipv4: 		  ipv4,
		ipv6:		  ipv6,
		nsqInTopic:       "zdns_4hr_results",
		nsqZDNSOutTopic:  "zdns_24hr",
		nsqZGrabOutTopic: "zgrab",
		zdnsDelay:        86400, // 24hours -- 3600 sec * 24
	}
	return NewSentinelZDNSOrchestrator(*cfg24hr)
}

func NewSentinelZDNSOrchestrator(cfg SentinelOrchestratorConfig) *SentinelZDNSOrchestrator {
	nsqHost := cfg.nsqHost
	ipv4 := cfg.ipv4
	ipv6 := cfg.ipv6
	// Instantiate a consumer that will subscribe to the provided channel.
	consumer, err := nsq.NewConsumer(cfg.nsqInTopic, "orchestrator", nsq.NewConfig())
	consumer.SetLoggerLevel(nsq.LogLevelError)
	if err != nil {
		log.Fatal(err)
	}
	// Create a new NSQ producer
	nsqUrl := fmt.Sprintf("%s:4150", nsqHost)
	producer, err := nsq.NewProducer(nsqUrl, nsq.NewConfig())
	producer.SetLoggerLevel(nsq.LogLevelError)
	if err != nil {
		// Report Error and Exit.
		log.Fatal(err)
	}

	return &SentinelZDNSOrchestrator{
		nsqHost:          nsqHost,
		ipv4: 		  ipv4,
		ipv6:		  ipv6,
		consumer:         *consumer,
		producer:         *producer,
		nsqZDNSOutTopic:  cfg.nsqZDNSOutTopic,
		nsqZGrabOutTopic: cfg.nsqZGrabOutTopic,
		zdnsDelay:        cfg.zdnsDelay,
	}
}

func (szo *SentinelZDNSOrchestrator) feedZDNSDelayed(metadata ZDNSMetadata, name string) error {
	scanAfter := metadata.ScanAfter
	newScanAfter, _ := strconv.ParseInt(scanAfter, 0, 64)
	newScanAfter = newScanAfter + szo.zdnsDelay

	// fmt.Printf("New Scan After: %d; Delay: %d", newScanAfter, szo.zdnsDelay)
	zdnsFeedInput := fmt.Sprintf("{\"domain\": \"%s\",\"metadata\": {\"cert_sha1\": \"%s\", \"scan_after\": \"%d\"}}", name, metadata.CertSHA1, newScanAfter)
	err := szo.producer.Publish(szo.nsqZDNSOutTopic, []byte(zdnsFeedInput))
	log.Info(fmt.Sprintf("ZDNS to 4/24hr: Publishing %s to channel %s", zdnsFeedInput, szo.nsqZDNSOutTopic))
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (szo *SentinelZDNSOrchestrator) feedZGrab(IPv4Addresses []string, IPv6Addresses []string, name string) error {
	if (szo.ipv4) {
		for _, ipv4 := range IPv4Addresses {
			tnow := time.Now().Unix()
			zgrabInput := fmt.Sprintf("{\"sni\": \"%s\", \"ip\": \"%s\",  \"scan_after\": \"%d\"}", name, ipv4, tnow)
			log.Info(fmt.Sprintf("ZDNS to Zgrab IPV4: Publishing %s to channel %s", zgrabInput, szo.nsqZDNSOutTopic))
			err := szo.producer.Publish(szo.nsqZGrabOutTopic, []byte(zgrabInput))
			if err != nil {
				log.Error(err)
			}
		}
	}
	if (szo.ipv6) {
		for _, ipv6 := range IPv6Addresses {
			tnow := time.Now().Unix()
			zgrabInput := fmt.Sprintf("{\"sni\": \"%s\", \"ip\": \"%s\",  \"scan_after\": \"%d\"}", name, ipv6, tnow)
			log.Info(fmt.Sprintf("ZDNS to Zgrab IPV6: Publishing %s to channel %s", zgrabInput, szo.nsqZDNSOutTopic))
			err := szo.producer.Publish(szo.nsqZGrabOutTopic, []byte(zgrabInput))
			if err != nil {
				log.Error(err)
			}
		}
	}
	return nil
}

func (szo *SentinelZDNSOrchestrator) FeedBroker() error {
	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	szo.consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		var Result ZDNSResult
		// handle the message
		err := json.Unmarshal(m.Body, &Result)
		if err != nil {
			log.Error(err)
			return err
		}
		err = szo.feedZDNSDelayed(Result.MetaData, Result.Data.Name)
		if err != nil {
			log.Error(err)
			return err
		}
		err = szo.feedZGrab(Result.Data.IPv4Addresses, Result.Data.IPv6Addresses, Result.Data.Name)
		if err != nil {
			log.Error(err)
			return err
		}

		return nil
	}))

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	nsqUrl := fmt.Sprintf("%s:4161", szo.nsqHost)
	err := szo.consumer.ConnectToNSQLookupd(nsqUrl)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Switch to a better model
	// done channel + ticker
	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Gracefully stop the consumer.
	szo.consumer.Stop()
	szo.producer.Stop()
	return nil
}
