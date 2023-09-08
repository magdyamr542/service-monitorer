package informer

import (
	"encoding/json"
	"fmt"
)

type SupportedInformer string

const (
	Slack SupportedInformer = "slack"
)

var (
	SupportedInformers = []SupportedInformer{Slack}
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

type Config struct {
	Name string            `yaml:"name"`
	Type SupportedInformer `yaml:"type"`
	// Configures the corresponding informer based on its type.
	Config map[string]interface{} `yaml:"config"`
}

func (i Config) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("name is required")
	}
	if i.Type == "" {
		return fmt.Errorf("type is required")
	}

	validType := false
	for _, supported := range SupportedInformers {
		if i.Type == supported {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("informer with type %s is not valid. supported types are %s", i.Type, SupportedInformers)
	}

	switch SupportedInformer(i.Type) {
	case Slack:
		jsonbody, err := json.Marshal(i.Config)
		if err != nil {
			return fmt.Errorf("slack config to json: %v", err)
		}
		slackConfig := SlackConfig{}
		if err := json.Unmarshal(jsonbody, &slackConfig); err != nil {
			return fmt.Errorf("slack config %+v invalid: %v", i.Config, err)
		}
		if err := slackConfig.Validate(); err != nil {
			return fmt.Errorf("slack config invalid: %v", err)
		}

	}

	return nil
}
