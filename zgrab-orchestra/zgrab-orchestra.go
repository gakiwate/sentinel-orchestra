package zgraborc

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	mon "github.com/gakiwate/sentinel-orchestra/sentinel-monitor"
	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"
)

type SentinelZGrabOrchestrator struct {
	monitor          *mon.SentinelMonitor
	nsqHost          string
	consumer         nsq.Consumer
	producer         nsq.Producer
	nsqZGrabOutTopic string
	zgrabDelay       int64
}

type ZGrabMetadata struct {
	CertSHA1  string `json:"cert_sha1"`
	ScanAfter string `json:"scan_after"`
	CertType  string `json:"cert_type"`
}

type ZGrabResult struct {
	IP       string        `json:"ip"`
	Domain   string        `json:"domain"`
	MetaData ZGrabMetadata `json:"metadata"`
}

type SentinelOrchestratorConfig struct {
	monitor          *mon.SentinelMonitor
	nsqHost          string
	nsqInTopic       string
	nsqZGrabOutTopic string
	zgrabDelay       int64
}

func NewSentinelZgrab4hrDelayOrchestrator(monitor *mon.SentinelMonitor, nsqHost string) *SentinelZGrabOrchestrator {
	cfg4hr := &SentinelOrchestratorConfig{
		monitor:          monitor,
		nsqHost:          nsqHost,
		nsqInTopic:       "zgrab_results",
		nsqZGrabOutTopic: "zgrab_4hr",
		zgrabDelay:       14400, // 4hours -- 3600 sec * 4
	}
	return NewSentinelZGrabOrchestrator(*cfg4hr)
}

func NewSentinelZgrab8hrDelayOrchestrator(monitor *mon.SentinelMonitor, nsqHost string) *SentinelZGrabOrchestrator {
	cfg8hr := &SentinelOrchestratorConfig{
		monitor:          monitor,
		nsqHost:          nsqHost,
		nsqInTopic:       "zgrab_4hr_results",
		nsqZGrabOutTopic: "zgrab_8hr",
		zgrabDelay:       28800, // 8hours -- 3600 sec * 8
	}
	return NewSentinelZGrabOrchestrator(*cfg8hr)
}

func NewSentinelZGrabOrchestrator(cfg SentinelOrchestratorConfig) *SentinelZGrabOrchestrator {
	nsqHost := cfg.nsqHost
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

	return &SentinelZGrabOrchestrator{
		monitor:          cfg.monitor,
		nsqHost:          nsqHost,
		consumer:         *consumer,
		producer:         *producer,
		nsqZGrabOutTopic: cfg.nsqZGrabOutTopic,
		zgrabDelay:       cfg.zgrabDelay,
	}
}

func (szo *SentinelZGrabOrchestrator) feedZGrabDelayed(metadata ZGrabMetadata, IP string, Domain string) error {
	ScanAfter := metadata.ScanAfter
	newScanAfter, _ := strconv.ParseInt(ScanAfter, 0, 64)
	newScanAfter = newScanAfter + szo.zgrabDelay
	zgrabInput := fmt.Sprintf("{\"sni\": \"%s\", \"ip\": \"%s\", \"metadata\": {\"scan_after\": \"%d\", \"cert_sha1\": \"%s\", \"cert_type\": \"%s\"}}", Domain, IP, newScanAfter, metadata.CertSHA1, metadata.CertType)

	err := szo.producer.Publish(szo.nsqZGrabOutTopic, []byte(zgrabInput))
	log.Info(fmt.Sprintf("Zgrab to 4/8hr: Publishing %s to channel %s", zgrabInput, szo.nsqZGrabOutTopic))

	if err != nil {
		log.Error(err)
	}

	return nil
}

func (szo *SentinelZGrabOrchestrator) FeedBroker() error {
	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	szo.consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		var Result ZGrabResult
		// handle the message
		err := json.Unmarshal(m.Body, &Result)
		if err != nil {
			log.Error(err)
			return err
		}
		szo.monitor.Stats.Incr("stats.zgrab.result_cnt")
		err = szo.feedZGrabDelayed(Result.MetaData, Result.IP, Result.Domain)

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
