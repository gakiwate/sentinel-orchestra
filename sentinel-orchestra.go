package main

import (
	"os"
	"os/signal"
	"syscall"

	zdnsorc "github.com/gakiwate/sentinel-orchestra/zdns-orchestra"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	var nsqHost string

	rootCmd := &cobra.Command{
		Use:   "sentinel-orchestra",
		Short: "Orchestrator to broker Sentinel messages",
		Long:  "sentinel-orchestrator manages messages between the different sentinel programs",
		Run: func(cmd *cobra.Command, args []string) {
			nsqHost, _ = cmd.Flags().GetString("nsq-host")
		},
	}

	rootCmd.Flags().StringVar(&nsqHost, "nsq-host", "localhost", "IP address of machine running nslookupd")

	// Set Logger Level
	log.SetLevel(log.ErrorLevel)

	// TODO: FIX weird behavior on --help / -h flag.
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	zdnsOrchestrator := zdnsorc.NewSentinelZDNS4hrDelayOrchestrator(nsqHost)
	go zdnsOrchestrator.FeedBroker()

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
