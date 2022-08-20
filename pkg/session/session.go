package session

import (
	"chrono/pkg/config"
	"chrono/pkg/event/periodic"
	"chrono/pkg/event/save"
	"chrono/pkg/repository"
	"chrono/pkg/scheduler"
	"chrono/pkg/signal"
	"context"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
)

type Session struct {
	r *repository.Repository
}

func NewSession(repositoryPath string) *Session {
	r := repository.Open(repositoryPath)
	log.Info().Str("repository", repositoryPath).Msg("GIT repository")

	head, err := r.Git.Head()
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	branch := head.Branch()

	branchName, err := branch.Name()
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	if branchName != "chrono" {
		log.Error().Msg("Please git checkout to a branch named \"chrono\" to proceed")
		log.Fatal().Str("branch", branchName).Msg("Currently not in a chrono branch")
	}

	return &Session{
		r: r,
	}
}

func (s *Session) Start() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-signal.Ch:
				cancel()
				wg.Wait()
				os.Exit(0)
			}
		}
	}()

	scheduler.Init(ctx)
	scheduler.SetRepository(s.r)

	wg.Add(1)
	go func() {
		defer wg.Done()
		scheduler.Run()
		scheduler.Fini()
	}()

	if config.Cfg.Events.Periodic != nil {
		scheduler.AddEvent(periodic.Periodic)
	}

	if config.Cfg.Events.Save != nil {
		scheduler.AddEvent(save.Save)
	}

	wg.Wait()
}
