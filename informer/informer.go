package informer

import (
	"context"
	"text/template"
)

type PingResult struct {
	Backend    string
	Status     string
	StatusCode int
	Timestamp  string
	Failures   []ServiceFailure
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
type ServiceFailure struct {
	// Name of the component that failed.
	Name string
	// ok or failed.
	Status string
	// The error that led to failure.
	Error string
	// Whether the error is fatal.
	Fatal bool
}

type Informer interface {
	Inform(ctx context.Context, config Config, backend string, pingResult PingResult, template *template.Template) error
}
