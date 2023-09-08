package informer

import "context"

type Informer interface {
	Inform(ctx context.Context, config Config, message string) error
}
