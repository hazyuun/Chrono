package repository

import (
	"chrono/pkg/config"
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

func (r *Repository) GetBranchName() string {
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

	return currentBranchName
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

	if config.Cfg.Git != nil && config.Cfg.Git.AutoAdd {
		log.Info().Msg("Auto-adding all files")
		index.AddAll([]string{"*"}, git.IndexAddCheckPathspec, func(s1, s2 string) error {
			updatesExist = true
			return nil
		})
	}

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

	r.Git.CheckoutHead(&git.CheckoutOptions{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing,
	})

	log.Info().Str("id", commitId.String()).Msg("New git commit")
}

func (r *Repository) SquashMerge(dst string, src string, msg string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Step 1: Checkout to destination branch
	branch, err := r.Git.LookupBranch(dst, git.BranchLocal)
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

	r.sessionBranch = dst

	// Step 2: Get both the destination and source branch for later use
	srcBranch, err := r.Git.LookupBranch(src, git.BranchLocal)
	if err != nil {
		log.Fatal().Err(err).Str("src", src).Msg("GIT Error, Failed to lookup source branch")
	}

	dstBranch, err := r.Git.LookupBranch(dst, git.BranchLocal)
	if err != nil {
		log.Fatal().Err(err).Str("destination", src).Msg("GIT Error, Failed to lookup destination branch")
	}

	// Step 3: Do merge analysis
	ac, err := r.Git.AnnotatedCommitFromRef(srcBranch.Reference)
	if err != nil {
		log.Fatal().Err(err).Str("src", src).Msg("GIT Error, Failed get annotated commit")
	}

	head, err := r.Git.Head()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to get HEAD")
	}

	mergeHeads := make([]*git.AnnotatedCommit, 1)
	mergeHeads[0] = ac
	analysis, _, err := r.Git.MergeAnalysis(mergeHeads)
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, Merge analysis failed")
	}

	if analysis&git.MergeAnalysisNone != 0 || analysis&git.MergeAnalysisUpToDate != 0 {
		log.Fatal().Msg("GIT Error, Nothing to merge")
	}

	if analysis&git.MergeAnalysisNormal == 0 {
		log.Fatal().Msg("GIT Error, Git merge analysis reported a not normal merge")
	}

	// Step 4: Merge
	mergeOpts, err := git.DefaultMergeOptions()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, DefaultMergeOptions() failed")
	}

	mergeOpts.FileFavor = git.MergeFileFavorNormal
	mergeOpts.TreeFlags = git.MergeTreeFailOnConflict

	checkoutOpts := &git.CheckoutOptions{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing | git.CheckoutUseTheirs,
	}

	err = r.Git.Merge(mergeHeads, &mergeOpts, checkoutOpts)
	if err != nil {
		log.Fatal().Err(err).Msg("GIT error, Merge failed")
	}

	index, err := r.Git.Index()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to retreive index")
	}

	if index.HasConflicts() {
		log.Fatal().Msg("GIT Error, Merge conflicts, please solve them and commit manually")
	}

	// Step 5: Commit
	commit, err = r.Git.LookupCommit(dstBranch.Target())
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to lookup commit")
	}

	sig := &git.Signature{
		Name:  "Chrono",
		Email: "Chrono",
		When:  time.Now(),
	}

	treeId, err := index.WriteTree()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, Failed to write tree")
	}

	t, err := r.Git.LookupTree(treeId)
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, Failed to lookup tree")
	}

	currentCommit, err := r.Git.LookupCommit(head.Target())
	if err != nil {
		log.Fatal().Err(err).Msg("GIT error, Failed to get current commit")
	}

	commitId, err := r.Git.CreateCommit("HEAD", sig, sig, msg, t, currentCommit)
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, failed to create commit")
	}

	r.Git.CheckoutHead(&git.CheckoutOptions{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing,
	})

	log.Info().Str("id", commitId.String()).Msg("New git commit")

	// Step 6: Cleanup the state
	err = r.Git.StateCleanup()
	if err != nil {
		log.Fatal().Err(err).Msg("GIT Error, Failed to cleanup state")
	}
}
