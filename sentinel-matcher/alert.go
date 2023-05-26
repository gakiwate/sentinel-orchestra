package matcher

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type Alert struct {
	ChannelID string
	APIKey    string
}

func NewAlert() *Alert {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	channelID := os.Getenv("SLACK_CHANNEL_ID")
	apiKey := os.Getenv("SLACK_API_KEY")

	return &Alert{
		ChannelID: channelID,
		APIKey:    apiKey,
	}
}

func (a *Alert) SendAlert(message string) {
	api := slack.New(a.APIKey)
	channelID, timestamp, err := api.PostMessage(
		a.ChannelID,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	log.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}
