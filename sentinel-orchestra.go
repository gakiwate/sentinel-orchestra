package main

import (
	"io/ioutil"

	certstreamorc "github.com/gakiwate/sentinel-orchestra/certstream-orchestra"
	sentinelmon "github.com/gakiwate/sentinel-orchestra/sentinel-monitor"
	zdnsorc "github.com/gakiwate/sentinel-orchestra/zdns-orchestra"
	zgraborc "github.com/gakiwate/sentinel-orchestra/zgrab-orchestra"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Certstream struct {
		Enable bool     `yaml:"enable"`
		Topics []string `yaml:"topics"`
	} `yaml:"certstream"`
	ZDNS struct {
		Enable bool     `yaml:"enable"`
		Topics []string `yaml:"topics"`
	} `yaml:"zdns"`
	ZGrab struct {
		Enable bool     `yaml:"enable"`
		Topics []string `yaml:"topics"`
	} `yaml:"zgrab"`
}

func main() {
	var nsqHost string
	var nsqOutTopic string

	rootCmd := &cobra.Command{
		Use:   "sentinel-orchestra",
		Short: "Orchestrator to broker Sentinel messages",
		Long:  "sentinel-orchestrator manages messages between the different sentinel programs",
		Run: func(cmd *cobra.Command, args []string) {
			nsqHost, _ = cmd.Flags().GetString("nsq-host")
			nsqOutTopic, _ = cmd.Flags().GetString("nsq-topic")
		},
	}

	rootCmd.Flags().StringVar(&nsqHost, "nsq-host", "localhost", "IP address of machine running nslookupd")
	rootCmd.Flags().StringVar(&nsqOutTopic, "nsq-topic", "zdns", "The NSQ topic to publish on")

	// Set Logger Level
	log.SetLevel(log.ErrorLevel)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	if rootCmd.Flags().Changed("help") {
		return
	}

	// Use configuration file to determine which programs to run
	configData, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	monitor := sentinelmon.NewSentinelMonitor()

	if config.Certstream.Enable {
		certstreamOrchestrator := certstreamorc.NewSentinelCertstreamOrchestrator(monitor, nsqHost, config.Certstream.Topics[0])
		go certstreamOrchestrator.Run()
	}

	if config.ZDNS.Enable {
		for _, topic := range config.ZDNS.Topics {
			if topic == "zdns_4hr" {
				zdnsOrchestrator_4hr := zdnsorc.NewSentinelZDNS4hrDelayOrchestrator(monitor, nsqHost)
				go zdnsOrchestrator_4hr.FeedBroker()
			}
			if topic == "zdns_24hr" {
				zdnsOrchestrator_24hr := zdnsorc.NewSentinelZDNS24hrDelayOrchestrator(monitor, nsqHost)
				go zdnsOrchestrator_24hr.FeedBroker()
			}
		}
	}

	if config.ZGrab.Enable {
		for _, topic := range config.ZGrab.Topics {
			if topic == "zgrab_4hr" {
				zgrabOrchestrator_4hr := zgraborc.NewSentinelZgrab4hrDelayOrchestrator(monitor, nsqHost)
				go zgrabOrchestrator_4hr.FeedBroker()
			}
			if topic == "zgrab_24hr" {
				zgrabOrchestrator_24hr := zgraborc.NewSentinelZgrab24hrDelayOrchestrator(monitor, nsqHost)
				go zgrabOrchestrator_24hr.FeedBroker()
			}
		}
	}

	log.Fatal(monitor.Serve())
}
