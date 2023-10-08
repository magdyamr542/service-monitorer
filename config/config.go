package config

import (
	"fmt"
	"text/template"

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
	Auth        *BackendAuth    `yaml:"auth"`
	Response    BackendResponse `yaml:"response"`
}

type BackendAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BackendResponse struct {
	ExpectCode int    `yaml:"expectCode"`
	OnFail     OnFail `yaml:"onFail"`
}

type OnFail struct {
	Inform []struct {
		// Name of the informer.
		Informer string `yaml:"informer"`
		Template string `yaml:"template"`
	} `yaml:"inform"`
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

	if b.Auth != nil {
		if err := b.Auth.Validate(); err != nil {
			return fmt.Errorf("backend auth invalid: %v", err)
		}
	}

	if err := b.Response.Validate(availableInformers); err != nil {
		return fmt.Errorf("backend response invalid: %v", err)
	}

	return nil
}

func (b BackendAuth) Validate() error {
	if b.Username == "" {
		return fmt.Errorf("username is required")
	}

	if b.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (b BackendResponse) Validate(availableInformers []string) error {
	for _, inform := range b.OnFail.Inform {
		// Check the informer exists.
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

		if inform.Template == "" {
			return fmt.Errorf("empty template for informer %s", inform.Informer)
		}
	}

	return nil
}

// TemplateMap maps the combination (backend name, informer name) to a go template that will be used
// to construct the message that is sent to the corresponding informer.
type TemplateMap map[string]*template.Template

func (c Config) InitTemplateMap() (TemplateMap, error) {
	tm := make(TemplateMap)
	for _, backend := range c.Backends {
		for _, informConfig := range backend.Response.OnFail.Inform {

			name := TemplateName(backend.Name, informConfig.Informer)
			if _, ok := tm[name]; ok {
				return nil, fmt.Errorf("error parsing template for backend %q, informer %q. can't be defined more than once",
					backend.Name, informConfig.Informer)
			}

			parsed, err := template.New(name).Parse(informConfig.Template)
			if err != nil {
				return nil, fmt.Errorf("error parsing template for backend %q, informer %q: %v",
					backend.Name, informConfig.Informer, err)
			}

			tm[name] = parsed
		}
	}

	return tm, nil
}

func TemplateName(backend, informer string) string {
	return backend + "." + informer
}
