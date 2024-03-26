package main

import (
	"fmt"
	"os"

	certstreamorc "github.com/gakiwate/sentinel-orchestra/certstream-orchestra"
	sentineldb "github.com/gakiwate/sentinel-orchestra/sentinel-db"
	sentinelmon "github.com/gakiwate/sentinel-orchestra/sentinel-monitor"
	zdnsorc "github.com/gakiwate/sentinel-orchestra/zdns-orchestra"
	zgraborc "github.com/gakiwate/sentinel-orchestra/zgrab-orchestra"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Certstream struct {
		Enable bool     `default:"false" yaml:"enable"`
		Topics []string `yaml:"topics"`
	} `yaml:"certstream"`
	ZDNS struct {
		Enable bool     `default:"false" yaml:"enable"`
		Ipv4   bool     `yaml:"ipv4"`
		Ipv6   bool     `yaml:"ipv6"`
		Topics []string `yaml:"topics"`
	} `yaml:"zdns"`
	ZGrab struct {
		Enable bool     `default:"false" yaml:"enable"`
		Topics []string `yaml:"topics"`
	} `yaml:"zgrab"`
	Monitor struct {
		StoragePath string `default:"." yaml:"storage"`
		Name        string `default:"sentinel-stats" yaml:"name"`
	} `yaml:"monitor"`
	DataStore struct {
		StoragePath string `default:"." yaml:"storage"`
	}
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
	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	monitorName := fmt.Sprintf("%s/%s", config.Monitor.StoragePath, config.Monitor.Name)
	monitor := sentinelmon.NewSentinelMonitor(monitorName)
	log.Info("Created the monitor")

	dbName := fmt.Sprintf("%s/%s", config.DataStore.StoragePath, "sentinel-data")
	db := sentineldb.NewSentinelDB(dbName, false)
	log.Info("Created Data Store")

	if config.Certstream.Enable {
		certstreamOrchestrator := certstreamorc.NewSentinelCertstreamOrchestrator(db, monitor, nsqHost, config.Certstream.Topics[0])
		go certstreamOrchestrator.Run()
	}
	log.Info("Launched certstream orchestrator")

	if config.ZDNS.Enable {
		ipv4 := config.ZDNS.Ipv4
		ipv6 := config.ZDNS.Ipv6
		for _, topic := range config.ZDNS.Topics {
			if topic == "zdns_4hr" {
				zdnsOrchestrator_4hr := zdnsorc.NewSentinelZDNS4hrDelayOrchestrator(db, monitor, nsqHost, ipv4, ipv6)
				go zdnsOrchestrator_4hr.FeedBroker()
			}
			if topic == "zdns_8hr" {
				zdnsOrchestrator_8hr := zdnsorc.NewSentinelZDNS8hrDelayOrchestrator(db, monitor, nsqHost, ipv4, ipv6)
				go zdnsOrchestrator_8hr.FeedBroker()
			}
		}
	}

	if config.ZGrab.Enable {
		for _, topic := range config.ZGrab.Topics {
			if topic == "zgrab_4hr" {
				zgrabOrchestrator_4hr := zgraborc.NewSentinelZgrab4hrDelayOrchestrator(monitor, nsqHost)
				go zgrabOrchestrator_4hr.FeedBroker()
			}
			if topic == "zgrab_8hr" {
				zgrabOrchestrator_8hr := zgraborc.NewSentinelZgrab8hrDelayOrchestrator(monitor, nsqHost)
				go zgrabOrchestrator_8hr.FeedBroker()
			}
		}
	}

	log.Fatal(monitor.Serve())
}
