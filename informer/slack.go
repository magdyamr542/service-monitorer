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

const (
	blockTypeSection = "section"
	textTypeMarkdown = "mrkdwn"
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

type slackSendJsonMessageRequest struct {
	Blocks []block `json:"blocks"`
}

type block struct {
	Type string    `json:"type"`
	Text blockText `json:"text"`
}

type blockText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewSlack(logger *log.Logger, httpClient http.Client) Informer {
	l := logger.With("informerType", "slack")
	return slack{logger: l, httpClient: httpClient}
}

func (s slack) Inform(ctx context.Context, config Config, backend string, pingResult PingResult) error {
	url := config.Config["webhookUrl"].(string)

	// s.logger.With("informer", config.Name).With("slackUrl", url).
	// 	Debugf("Informer message to deliver:\n%s", message)

	message := s.buildMessage(pingResult, backend)
	requestBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	response, statusCode, err := s.httpClient.Post(url, map[string]string{"Content-Type": "application/json"}, bytes.NewReader(requestBytes))
	if err != nil {
		return err
	}

	if statusCode != gohttp.StatusOK {
		s.logger.With("responseMessage", string(response)).With("responseCode", statusCode).
			Warnf("Non ok response from slack")

		return fmt.Errorf("response code from slack %d. Response: %s", statusCode, string(response))
	}

	return nil
}

func (s slack) buildMessage(pingResult PingResult, backend string) slackSendJsonMessageRequest {
	message := slackSendJsonMessageRequest{Blocks: make([]block, 0)}

	// Add the header. Info about the backend + Status code
	message.Blocks = append(message.Blocks, block{
		Type: blockTypeSection,
		Text: blockText{
			Type: textTypeMarkdown,
			Text: fmt.Sprintf("Backend *%s* has problems.\nCode: %d.\nTime: %s",
				backend, pingResult.StatusCode, pingResult.Timestamp),
		},
	},
	)

	// Add the failures
	for _, failure := range pingResult.Failures {
		fatalStr := ""
		if failure.Fatal {
			fatalStr = "*FATAL*"
		}
		b := block{
			Type: blockTypeSection,
			Text: blockText{
				Type: textTypeMarkdown,
				Text: fmt.Sprintf("*%s* %q %s", failure.Name, failure.Reason, fatalStr),
			},
		}
		message.Blocks = append(message.Blocks, b)
	}
	return message
}
