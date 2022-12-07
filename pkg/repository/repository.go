package repository

import (
	"sync"
	"time"

	git "github.com/libgit2/git2go/v34"
	"github.com/rs/zerolog/log"
)

type Repository struct {
	Git           *git.Repository
	sessionBranch string
	mutex         sync.Mutex
}

func Open(path string) *Repository {
	r, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to open GIT repository")
	}

	return &Repository{
		Git: r,
	}
}

func (r *Repository) CreateBranch(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	head, err := r.Git.Head()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to get HEAD")
	}

	commit, err := r.Git.LookupCommit(head.Target())
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to get current commit")
	}

	log.Info().Str("commit", commit.Message()).Msg("Branching from commit")
	_, err = r.Git.CreateBranch(name, commit, false)
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to create branch")
	}
}

func (r *Repository) DeleteBranch(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	branch, err := r.Git.LookupBranch(name, git.BranchLocal)
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to lookup branch")
	}

	err = branch.Delete()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to delete branch")
	}
}

func (r *Repository) CheckoutBranch(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	branch, err := r.Git.LookupBranch(name, git.BranchLocal)
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to lookup branch")
	}

	commit, err := r.Git.LookupCommit(branch.Target())
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to get last commit")
	}

	tree, err := commit.Tree()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to retreive tree")
	}

	err = r.Git.CheckoutTree(tree, &git.CheckoutOptions{Strategy: git.CheckoutSafe})
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to checkout tree")
	}

	err = r.Git.SetHead(branch.Reference.Name())
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to set HEAD")
	}

	r.sessionBranch = name
}

func (r *Repository) AssertBranchNotChanged() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	head, err := r.Git.Head()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to get HEAD")
	}

	branch := head.Branch()
	currentBranchName, err := branch.Name()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to get branch name")
	}

	if r.sessionBranch != currentBranchName {
		log.Fatal().Str("expected", r.sessionBranch).
			Str("found", currentBranchName).
			Msg("Branch changed ! Please make sure the branch doesn't get changed while chrono is running")
	}
}

func (r *Repository) Commit(paths []string, message string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	head, err := r.Git.Head()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to get HEAD")
	}

	branch := head.Branch()

	index, err := r.Git.Index()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to retreive index")
	}

	updatesExist := false
	err = index.UpdateAll(paths, func(s1, s2 string) error {
		updatesExist = true
		return nil
	})

	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to update index")
	}

	if !updatesExist {
		log.Info().Msg("Didn't commit, There are no updates")
		return
	}

	err = index.Write()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to write index")
	}

	oid, err := index.WriteTree()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to write tree")
	}

	tree, err := r.Git.LookupTree(oid)
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to lookup tree")
	}

	lastCommit, err := r.Git.LookupCommit(branch.Target())
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to lookup commit")
	}

	sig := &git.Signature{
		Name:  "Chrono",
		Email: "Chrono",
		When:  time.Now(),
	}

	commitId, err := r.Git.CreateCommit("HEAD", sig, sig, message, tree, lastCommit)
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to create commit")
	}

	r.Git.CheckoutHead(&git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing,
	})

	log.Info().Str("id", commitId.String()).Msg("New git commit")
}
