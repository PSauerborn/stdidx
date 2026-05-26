package main

import (
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// GitCloner clones a remote standards repository into a local directory. The
// interface exists so tests can substitute a filesystem-only implementation
// in place of an actual git clone.
type GitCloner interface {
	Clone(repo GitRepository) error
}

// ExecGitCloner is the production GitCloner implementation. It shells out to
// the local `git` binary, which must be present on PATH.
type ExecGitCloner struct{}

// Clone shells out to `git clone` to fetch repo.Repository into repo.ClonePath.
// If repo.Branch is set the corresponding branch is checked out; if repo.Tag
// is set the corresponding tag is checked out. The caller is responsible for
// ensuring at most one of Branch/Tag is set. stdout and stderr of the git
// process are streamed to the parent process.
func (g *ExecGitCloner) Clone(repo GitRepository) error {
	log.WithFields(log.Fields{
		"url":    repo.Repository,
		"path":   repo.ClonePath,
		"branch": repo.Branch,
		"tag":    repo.Tag,
	}).Info("cloning git repository")

	args := []string{"clone", repo.Repository, repo.ClonePath}
	if repo.Branch != "" {
		args = append(args, "--branch", repo.Branch)
	}
	if repo.Tag != "" {
		args = append(args, "--tag", repo.Tag)
	}

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
