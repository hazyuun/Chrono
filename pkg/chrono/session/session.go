package session

import (
	"chrono/pkg/chrono"
	"chrono/pkg/config"
	"chrono/pkg/event/periodic"
	"chrono/pkg/event/save"
	"chrono/pkg/repository"
	"chrono/pkg/scheduler"
	"chrono/pkg/signal"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type SessionDef struct {
	Name   string `json: "Name"`
	Branch string `json: "Branch"`
	Source string `json: "Source"`
}

type Session struct {
	Info SessionDef
	r    *repository.Repository
}

func GetSessions() map[string]SessionDef {
	sessionsPath := filepath.Join(chrono.RootPath, chrono.DotChronoDirName, chrono.SessionsFileName)

	file, err := os.ReadFile(sessionsPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	sessions := make(map[string]SessionDef)

	err = json.Unmarshal(file, &sessions)
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	return sessions
}

func OpenSession(name string) *Session {
	r := repository.Open(chrono.RootPath)
	log.Info().Str("repository", chrono.RootPath).Msg("Opened GIT repository")

	info, ok := GetSessions()[name]
	if !ok {
		log.Fatal().Str("name", name).Msg("Session of that name doesn't exist")
	}

	return &Session{
		Info: info,
		r:    r,
	}
}

func CreateSession(name string) {
	r := repository.Open(chrono.RootPath)
	log.Info().Str("repository", chrono.RootPath).Msg("Opened GIT repository")

	sessions := GetSessions()

	if _, ok := sessions[name]; ok {
		log.Fatal().Str("name", name).Msg("Session of that name already exists")
	}

	var sb strings.Builder
	sb.WriteString("chrono/")
	sb.WriteString(name)
	branchName := sb.String()

	r.CreateBranch(branchName)

	sessions[name] = SessionDef{
		Name:   name,
		Branch: branchName,
		Source: r.GetBranchName(),
	}

	sessionsPath := filepath.Join(chrono.RootPath, chrono.DotChronoDirName, chrono.SessionsFileName)
	bytes, err := json.Marshal(&sessions)
	if err != nil {
		r.DeleteBranch(branchName)
		log.Fatal().Err(err).Msg("Marshal error")
	}

	err = os.WriteFile(sessionsPath, bytes, os.ModePerm)
	if err != nil {
		r.DeleteBranch(branchName)
		log.Fatal().Err(err).Msg("Error")
	}
}

func DeleteSession(name string) {
	r := repository.Open(chrono.RootPath)
	log.Info().Str("repository", chrono.RootPath).Msg("Opened GIT repository")

	sessions := GetSessions()

	s, ok := sessions[name]

	if !ok {
		log.Fatal().Str("name", name).Msg("Session of that name doesn't exist")
	}

	r.DeleteBranch(s.Branch)
	delete(sessions, name)

	sessionsPath := filepath.Join(chrono.RootPath, chrono.DotChronoDirName, chrono.SessionsFileName)
	bytes, err := json.Marshal(&sessions)
	if err != nil {
		log.Fatal().Err(err).Msg("Marshal error")
	}

	err = os.WriteFile(sessionsPath, bytes, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
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

	s.r.CheckoutBranch(s.Info.Branch)

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
