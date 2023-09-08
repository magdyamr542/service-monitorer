package informer

import (
	"context"

	"github.com/charmbracelet/log"
)

type slack struct {
	logger *log.Logger
}

func NewSlack(logger *log.Logger) Informer {
	l := logger.With("informer", "slack")
	return slack{logger: l}
}

func (s slack) Inform(ctx context.Context, config Config, message string) error {
	s.logger.With("name", config.Name).
		Debugf("%s", message)
	return nil
}
