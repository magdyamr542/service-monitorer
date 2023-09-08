package config

import (
	"fmt"

	"github.com/magdyamr542/service-monitorer/informer"
)

type Config struct {
	Name      string            `yaml:"name"`
	Informers []informer.Config `yaml:"informers"`
	Backends  []Backend         `yaml:"backends"`
}

type Backend struct {
	Name        string          `yaml:"name"`
	URL         string          `yaml:"url"`
	CallEachSec int             `yaml:"callEachSec"`
	Response    BackendResponse `yaml:"response"`
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

func (c Config) GetInformer(name string) (informer.Config, error) {
	for _, inf := range c.Informers {
		if inf.Name == name {
			return inf, nil
		}
	}

	return informer.Config{}, fmt.Errorf("no such informer found")
}

func (b Backend) Validate(availableInformers []string) error {
	if b.Name == "" {
		return fmt.Errorf("name is required")
	}

	if b.URL == "" {
		return fmt.Errorf("url is required")
	}

	if b.CallEachSec <= 0 {
		return fmt.Errorf("callEachSec can't be <= 0")
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
