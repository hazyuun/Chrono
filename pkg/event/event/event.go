package event

import "context"

type Event interface {
	Init(ctx context.Context) error
	Watch() error
	Fini() error
}
