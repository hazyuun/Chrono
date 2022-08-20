package save

import (
	"chrono/pkg/config"
	"chrono/pkg/scheduler"
	"context"
	"time"

	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type SaveEvent struct {
	ctx context.Context
}

var Save SaveEvent

var watcher *fsnotify.Watcher

func (event SaveEvent) Init(ctx context.Context) error {
	log.Info().
		Strs("files", config.Cfg.Events.Save.Files).
		Msg("Initializing Save")

	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	for _, file := range config.Cfg.Events.Save.Files {
		err = watcher.Add(file)
		if err != nil {
			return fmt.Errorf("Couldn't add %v : %v", file, err.Error())
		}
	}

	Save.ctx = ctx

	return nil
}

func (event SaveEvent) Watch() error {
	for {
		select {
		case <-Save.ctx.Done():
			return nil

		case e, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("Couldn't read watcher event")
			}

			if e.Op&fsnotify.Write == fsnotify.Write {
				scheduler.Notify(scheduler.SchedulerMessage{
					Sender:  "Save",
					Message: fmt.Sprintf("[Save] Updated %v %v", e.Name, time.Now().Format("15:04:05 02/01/2006")),
					Paths:   config.Cfg.Events.Save.Files,
				})
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("Couldn't read watcher error")
			}

			return fmt.Errorf("Watcher error: %v", err.Error())
		}

	}

	return nil
}

func (event SaveEvent) Fini() error {
	log.Info().Msg("Save stopped")
	return watcher.Close()
}
