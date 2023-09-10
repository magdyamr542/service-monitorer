package monitorer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	config     config.Config
	logger     *log.Logger
	httpClient http.Client
	informers  map[informer.SupportedInformer]informer.Informer
}

func NewMonitorer(config config.Config, httpClient http.Client,
	informers map[informer.SupportedInformer]informer.Informer, logger *log.Logger) Monitorer {
	m := monitorer{config: config, logger: logger, httpClient: httpClient, informers: informers}
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
				m.logger.With("err", err).Errorf("Can't ping backend %s.", backend.Name)
				continue
			}

			message := m.buildInformerMessage(backend, result)
			if err := m.informForBackend(ctx, backend, message); err != nil {
				m.logger.With("err", err).Errorf("Error informing backend %s", backend.Name)
			}

		case <-doneCh:
			return nil
		}
	}
}

func (m *monitorer) buildInformerMessage(backend config.Backend, result pingResult) string {
	var strb strings.Builder
	strb.WriteString(fmt.Sprintf("Backend: %s\n", backend.Name))
	strb.WriteString(fmt.Sprintf("StatusCode: %d\n", result.statusCode))
	strb.WriteString("Errors:\n")
	tab := "	"
	for _, failure := range result.failures {
		fatal := ""
		if failure.fatal {
			fatal = "(FATAL)"
		}
		strb.WriteString(fmt.Sprintf("%s - %s  %q %s\n", tab, failure.name, failure.reason, fatal))
	}
	return strb.String()
}

func (m *monitorer) informForBackend(ctx context.Context, backend config.Backend, message string) error {
	errs := make([]error, 0)
	for _, i := range backend.Response.OnFail.Inform {
		informerConfig, err := m.config.GetInformer(i.Informer)
		if err != nil {
			return err
		}

		informer := m.informers[informerConfig.Type]

		if err := informer.Inform(ctx, informerConfig, message); err != nil {
			m.logger.With("err", err).
				Warnf("Error informing %s. Will continue to inform any other possible informers...", informerConfig.Name)
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type pingResult struct {
	statusCode int
	failures   []serviceFailure
}

// E.g response:
//
//	   {
//		  "name": "pdfgen",
//		  "status": "failed",
//		  "error": "Status check failed: the returned status code 401 differ from the configured one: 200",
//		  "fatal": true
//		},
//
// name will be 'pdfgen'
// reason will be 'pdfgen'
type serviceFailure struct {
	// name of the service component that failed
	name string
	// Status check failed: the returned status code 401 differ from the configured one: 200
	reason string
	fatal  bool
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

func (m *monitorer) pingBackend(ctx context.Context, backend config.Backend) (pingResult, error) {
	m.logger.Debugf("Pinging backend %s on %s", backend.Name, backend.URL)
	response, code, err := m.httpClient.Get(backend.URL, nil)
	if err != nil {
		return pingResult{}, err
	}

	var backendResponse backendResponse
	if err := json.Unmarshal(response, &backendResponse); err != nil {
		return pingResult{}, err
	}

	failures := make([]serviceFailure, 0)
	for _, componentStatus := range backendResponse.Details {
		if componentStatus.Status == ok {
			continue
		}
		failures = append(failures, serviceFailure{
			name:   componentStatus.Name,
			reason: componentStatus.Error,
			fatal:  componentStatus.Fatal,
		})
	}

	return pingResult{
		statusCode: code,
		failures:   failures,
	}, nil
}
