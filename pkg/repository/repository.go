package repository

import (
	"chrono/pkg/chrono"
	"sync"
	"time"

	git "github.com/libgit2/git2go/v33"
	"github.com/rs/zerolog/log"
)

type Repository struct {
	Chrono *chrono.Chrono
	Git    *git.Repository
	mutex  sync.Mutex
}

func Open(path string) *Repository {
	r, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	return &Repository{
		Chrono: &chrono.Chrono{},
		Git:    r,
	}
}

func (r *Repository) Commit(paths []string, message string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
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
		log.Error().Msg("Current branch is not \"chrono\" anymore. Please make sure you don't change the branch while chrono is running")
		log.Fatal().Str("branch", branchName).Msg("Current branch is not \"chrono\" anymore")
	}

	index, err := r.Git.Index()
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	updatesExist := false
	err = index.UpdateAll(paths, func(s1, s2 string) error {
		updatesExist = true
		return nil
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	if !updatesExist {
		log.Info().Msg("Didn't commit, There are no updates")
		return
	}

	err = index.Write()
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	oid, err := index.WriteTree()
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	tree, err := r.Git.LookupTree(oid)
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	lastCommit, err := r.Git.LookupCommit(branch.Target())
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	sig := &git.Signature{
		Name:  "Chrono",
		Email: "Chrono",
		When:  time.Now(),
	}

	commitId, err := r.Git.CreateCommit("HEAD", sig, sig, message, tree, lastCommit)
	if err != nil {
		log.Fatal().Err(err).Msg("Error")
	}

	r.Git.CheckoutHead(&git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing,
	})

	log.Info().Str("id", commitId.String()).Msg("New git commit")
}
