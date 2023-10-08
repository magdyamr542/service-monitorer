package informer

import (
	"context"
	"text/template"
)

type PingResult struct {
	Backend    string
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
	// name of the service component that failed
	Name string
	// Status check failed: the returned status code 401 differ from the configured one: 200
	Reason string
	Fatal  bool
}

type Informer interface {
	Inform(ctx context.Context, config Config, backend string, pingResult PingResult, template *template.Template) error
}
