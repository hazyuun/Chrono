package scheduler

import (
	"chrono/pkg/event/event"
	"chrono/pkg/repository"
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

type SchedulerMessage struct {
	Sender  string
	Message string
	Paths   []string
}

var scheduler struct {
	repository *repository.Repository
	channel    chan SchedulerMessage
	eventsWG   sync.WaitGroup
	ctx        context.Context
	mutex      sync.Mutex
}

func Init(ctx context.Context) {
	scheduler.channel = make(chan SchedulerMessage)
	scheduler.ctx = ctx
	log.Info().Msg("Scheduler: Starting..")
}

func Fini() {
	log.Info().Msg("Scheduler: Stopping..")
	scheduler.eventsWG.Wait()
	close(scheduler.channel)
}

func AddEvent(event event.Event) {
	err := event.Init(scheduler.ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	scheduler.eventsWG.Add(1)

	go func() {
		defer scheduler.eventsWG.Done()
		event.Watch()
		event.Fini()
	}()
}

func SetRepository(r *repository.Repository) {
	scheduler.repository = r
}

func Notify(msg SchedulerMessage) {
	scheduler.channel <- msg
}

func Run() {
	if scheduler.repository == nil {
		log.Fatal().Msg("Can not start scheduler with a nil repository")
	}

	log.Info().Msg("Scheduler: Running")
	for {
		select {
		case <-scheduler.ctx.Done():
			return
		case msg := <-scheduler.channel:
			log.Info().Str("event", msg.Sender).Str("msg", msg.Message).Msg("Event")

			scheduler.repository.AssertBranchNotChanged()
			scheduler.repository.Commit(msg.Paths, msg.Message)
		}
	}
}
