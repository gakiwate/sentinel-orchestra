package main

import (
	"os"
	"os/signal"
	"syscall"

	certstreamorc "github.com/gakiwate/sentinel-orchestra/certstream-orchestra"
	zdnsorc "github.com/gakiwate/sentinel-orchestra/zdns-orchestra"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

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
	certstreamOrchestrator := certstreamorc.NewSentinelCertstreamOrchestrator(nsqHost, nsqOutTopic)
	go certstreamOrchestrator.Run()

	zdnsOrchestrator := zdnsorc.NewSentinelZDNS4hrDelayOrchestrator(nsqHost)
	go zdnsOrchestrator.FeedBroker()

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
