package periodic

import (
	"chrono/pkg/config"
	"chrono/pkg/scheduler"
	"context"
	"fmt"

	"time"

	"github.com/rs/zerolog/log"
)

type PeriodicEvent struct {
	ticker *time.Ticker
	ctx    context.Context
}

var Periodic PeriodicEvent

var ticker *time.Ticker

func (event PeriodicEvent) Init(ctx context.Context) error {
	log.Info().
		Int("period", config.Cfg.Events.Periodic.Period).
		Strs("files", config.Cfg.Events.Periodic.Files).
		Msg("Initializing Periodic")

	Periodic.ticker = time.NewTicker(time.Duration(config.Cfg.Events.Periodic.Period) * time.Second)
	Periodic.ctx = ctx
	return nil
}

func (event PeriodicEvent) Watch() error {
	for {
		select {
		case <-Periodic.ctx.Done():
			return nil

		case <-Periodic.ticker.C:
			scheduler.Notify(scheduler.SchedulerMessage{
				Sender:  "Periodic",
				Message: fmt.Sprintf("[Periodic] %v", time.Now().Format("15:04:05 02/01/2006")),
				Paths:   config.Cfg.Events.Periodic.Files,
			})
		}
	}

	return nil
}

func (event PeriodicEvent) Fini() error {
	log.Info().Msg("Periodic stopped")
	return nil
}
