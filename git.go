package main

import (
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type GitCloner interface {
	Clone(repo GitRepository) error
}

type ExecGitCloner struct{}

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
