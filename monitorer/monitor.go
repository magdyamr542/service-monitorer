package monitorer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/log"

	"github.com/magdyamr542/service-monitorer/config"
	"github.com/magdyamr542/service-monitorer/http"
	"github.com/magdyamr542/service-monitorer/informer"
)

type Monitorer interface {
	Monitor(context.Context) error
}

type monitorer struct {
	config      config.Config
	logger      *log.Logger
	httpClient  http.Client
	informers   map[informer.SupportedInformer]informer.Informer
	templateMap config.TemplateMap
}

func NewMonitorer(
	config config.Config,
	httpClient http.Client,
	informers map[informer.SupportedInformer]informer.Informer,
	logger *log.Logger,
	templateMap config.TemplateMap,
) Monitorer {
	m := monitorer{config: config,
		logger:      logger,
		httpClient:  httpClient,
		informers:   informers,
		templateMap: templateMap,
	}
	return &m
}

func (m *monitorer) Monitor(ctx context.Context) error {
	wg := sync.WaitGroup{}
	errors := make([]error, 0)

	// Monitor each backend in a go routine.
	for _, backend := range m.config.Backends {
		wg.Add(1)

		go func(backend config.Backend) {

			if err := m.monitorBackend(ctx, backend); err != nil {
				errors = append(errors, err)
			}
			wg.Done()

		}(backend)
	}

	wg.Wait()
	return nil
}

func (m *monitorer) monitorBackend(ctx context.Context, backend config.Backend) error {
	m.logger.With("backendUrl", backend.URL).
		Debugf("Will ping backend %q each %d seconds", backend.Name, backend.CallEachSec)

	doneCh := ctx.Done()
	ticker := time.NewTicker(time.Duration(backend.CallEachSec) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			result, err := m.pingBackend(ctx, backend)
			if err != nil {
				m.logger.With("err", err).Errorf("Can't ping backend %s", backend.Name)
				continue
			}

			if err := m.informForBackend(ctx, backend, result); err != nil {
				m.logger.With("err", err).Errorf("Error informing backend %s", backend.Name)
			}

		case <-doneCh:
			return nil
		}
	}
}

func (m *monitorer) informForBackend(ctx context.Context, backend config.Backend, pingResult informer.PingResult) error {
	errs := make([]error, 0)
	for _, i := range backend.Response.OnFail.Inform {
		informerConfig, err := m.config.GetInformer(i.Informer)
		if err != nil {
			return err
		}

		informer := m.informers[informerConfig.Type]
		template := m.templateMap[config.TemplateName(backend.Name, i.Informer)]
		if err := informer.Inform(ctx, informerConfig, backend.Name, pingResult, template); err != nil {
			m.logger.With("err", err).
				Warnf("Error informing %s. Will continue to inform any other possible informers...", informerConfig.Name)
			errs = append(errs, fmt.Errorf("error informing %s: %v", i.Informer, err))
		}
	}
	return errors.Join(errs...)
}

type componentStatus string

const (
	ok     componentStatus = "ok"
	failed componentStatus = "failed"
)

// backendResponse is the interface for the response of all backend services
type backendResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Details   []struct {
		Name   string          `json:"name"`
		Status componentStatus `json:"status"`
		Error  string          `json:"error"`
		Fatal  bool            `json:"fatal"`
	} `json:"details"`
}

func (m *monitorer) pingBackend(ctx context.Context, backend config.Backend) (informer.PingResult, error) {
	m.logger.Debugf("Pinging backend %s on %s", backend.Name, backend.URL)

	var auth *http.BasicAuth
	if backend.Auth != nil {
		auth = &http.BasicAuth{Username: backend.Auth.Username, Password: backend.Auth.Password}
	}
	response, code, err := m.httpClient.Get(backend.URL, nil, auth)
	if err != nil {
		return informer.PingResult{}, err
	}

	var backendResponse backendResponse
	if err := json.Unmarshal(response, &backendResponse); err != nil {
		return informer.PingResult{}, err
	}

	failures := make([]informer.ServiceFailure, 0)
	for _, componentStatus := range backendResponse.Details {
		if componentStatus.Status == ok {
			continue
		}
		failures = append(failures, informer.ServiceFailure{
			Name:   componentStatus.Name,
			Reason: componentStatus.Error,
			Fatal:  componentStatus.Fatal,
		})
	}

	return informer.PingResult{
		Backend:    backend.Name,
		StatusCode: code,
		Failures:   failures,
		Timestamp:  backendResponse.Timestamp,
	}, nil
}
