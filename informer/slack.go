package informer

import (
	"bytes"
	"context"
	"fmt"
	gohttp "net/http"
	"text/template"

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

func NewSlack(logger *log.Logger, httpClient http.Client) Informer {
	l := logger.With("informerType", "slack")
	return slack{logger: l, httpClient: httpClient}
}

func (s slack) Inform(ctx context.Context, config Config, backend string, pingResult PingResult, template *template.Template) error {
	url := config.Config["webhookUrl"].(string)

	var buf bytes.Buffer
	err := template.Execute(&buf, pingResult)
	if err != nil {
		return fmt.Errorf("can't execute go template for backend %s: %v", backend, err)
	}

	response, statusCode, err := s.httpClient.Post(url, map[string]string{"Content-Type": "application/json"}, &buf, nil)
	if err != nil {
		return err
	}

	if statusCode != gohttp.StatusOK {
		return fmt.Errorf("response code from slack %d. Response: %s", statusCode, string(response))
	}

	return nil
}
