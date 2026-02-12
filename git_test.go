package main

import (
	log "github.com/sirupsen/logrus"
)

type MockGitCloner struct {
	SourceDir string
}

func (m *MockGitCloner) Clone(repo GitRepository) error {
	log.WithFields(log.Fields{
		"clonePath": repo.ClonePath,
		"sourceDir": m.SourceDir,
	}).Debug("cloning repository")

	return copyDir(m.SourceDir, repo.ClonePath)
}

func NewMockGitCloner(dir string) *MockGitCloner {
	return &MockGitCloner{
		SourceDir: dir,
	}
}
