package informer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	gohttp "net/http"

	"github.com/charmbracelet/log"
	"github.com/magdyamr542/service-monitorer/http"
)

type SlackConfig struct {
	WebhookURL string `yaml:"webhookUrl"`
}

func (s SlackConfig) Validate() error {
	if s.WebhookURL == "" {
		return fmt.Errorf("webhookUrl is required")
	}
	return nil
}

type slack struct {
	logger     *log.Logger
	httpClient http.Client
}

type slackSendMessageRequest struct {
	Text string `json:"text"`
}

type slackSendJsonMessageRequest struct {
	Blocks []block `json:"blocks"`
}

type block struct {
	Type string `json:"type"`
	Text struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"text"`
}

func NewSlack(logger *log.Logger, httpClient http.Client) Informer {
	l := logger.With("informerType", "slack")
	return slack{logger: l, httpClient: httpClient}
}

func (s slack) Inform(ctx context.Context, config Config, message string) error {
	url := config.Config["webhookUrl"].(string)

	s.logger.With("informer", config.Name).With("slackUrl", url).
		Debugf("Informer message to deliver:\n%s", message)

	request := slackSendMessageRequest{Text: message}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}
	_, statusCode, err := s.httpClient.Post(url, map[string]string{"Content-Type": "application/json"}, bytes.NewReader(requestBytes))
	if err != nil {
		return err
	}

	if statusCode != gohttp.StatusOK {
		return fmt.Errorf("response code from slack %d", statusCode)
	}

	return nil
}
