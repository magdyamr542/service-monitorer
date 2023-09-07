package config

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

type Config struct {
	Name      string     `yaml:"name"`
	Informers []Informer `yaml:"informers"`
	Backends  []Backend  `yaml:"backends"`
}

type SlackConfig struct {
	WebhookURL string `yaml:"webhookUrl"`
}

type Informer struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	// Configures the corresponding informer based on its type.
	Config map[string]interface{} `yaml:"config"`
}

type Backend struct {
	Name     string          `yaml:"name"`
	URL      string          `yaml:"url"`
	Response BackendResponse `yaml:"response"`
}

type BackendResponse struct {
	ExpectCode int `yaml:"expectCode"`
	OnFail     struct {
		Inform []struct {
			Informer string `yaml:"informer"`
			Template string `yaml:"template"`
		} `yaml:"inform"`
	} `yaml:"onFail"`
}

func (c Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	availableInformers := make([]string, 0)
	for _, inf := range c.Informers {
		if err := inf.Validate(); err != nil {
			return fmt.Errorf("informer %s invalid: %v", inf.Name, err)
		}
		availableInformers = append(availableInformers, inf.Name)
	}

	for _, b := range c.Backends {
		if err := b.Validate(availableInformers); err != nil {
			return fmt.Errorf("backend %s invalid: %v", b.Name, err)
		}
	}

	return nil
}

func (i Informer) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("name is required")
	}
	if i.Type == "" {
		return fmt.Errorf("type is required")
	}

	validType := false
	for _, supported := range SupportedInformers {
		if SupportedInformer(i.Type) == supported {
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

func (b Backend) Validate(availableInformers []string) error {
	if b.Name == "" {
		return fmt.Errorf("name is required")
	}

	if b.URL == "" {
		return fmt.Errorf("url is required")
	}

	if err := b.Response.Validate(availableInformers); err != nil {
		return fmt.Errorf("backend response invalid: %v", err)
	}

	return nil
}

func (b BackendResponse) Validate(availableInformers []string) error {
	for _, inform := range b.OnFail.Inform {
		if inform.Template == "" {
			return fmt.Errorf("template is required")
		}
		// Check the informer exists
		exists := false
		for _, informer := range availableInformers {
			if informer == inform.Informer {
				exists = true
				break
			}
		}

		if !exists {
			return fmt.Errorf("informer %s doesn't exist", inform.Informer)
		}
	}

	return nil
}

func (s SlackConfig) Validate() error {
	if s.WebhookURL == "" {
		return fmt.Errorf("webhookUrl is required")
	}
	return nil
}
