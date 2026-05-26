package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

// DefaultClonePath returns the default directory standards are cloned into.
// It resolves to "$HOME/.stdidx" when the home directory can be determined,
// and falls back to the relative ".stdidx" if it cannot.
func DefaultClonePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Warn("could not resolve home directory; falling back to relative path")
		return DefaultCloneDirName
	}
	return filepath.Join(home, DefaultCloneDirName)
}

// ExpandPath expands a leading "~" or "~/" in p to the user's home directory.
// Other path forms (including "~user/...") are returned unchanged.
func ExpandPath(p string) string {
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	return filepath.Join(home, strings.TrimPrefix(p, "~"))
}

// PrintSuggestedInstructions prints suggested agent instructions referencing
// the absolute path to the generated standards tree.
func PrintSuggestedInstructions(treePath string) {
	println("\nDon't forget to instruct your agent to use the standards index. Suggested prompt:")
	println("\n" + fmt.Sprintf(SuggestedAgentInstructionsTemplate, treePath) + "\n")
}

// Sync clones the remote standards repository described by repository into
// repository.ClonePath (removing any existing contents at that location
// first) and writes the generated standards tree to repository.ClonePath +
// "/" + TreeFileName. The ctx parameter is accepted for symmetry with the
// CLI command surface but is not currently propagated to git or filesystem
// operations.
func Sync(ctx context.Context, cloner GitCloner, repository GitRepository) error {
	log.WithFields(log.Fields{
		"repository": repository.Repository,
		"branch":     repository.Branch,
		"tag":        repository.Tag,
	}).Info("syncing standards library")

	if _, err := os.Stat(repository.ClonePath); err == nil {
		log.WithFields(log.Fields{
			"clone_path": repository.ClonePath,
		}).Info("removing existing standards library")
		if err := os.RemoveAll(repository.ClonePath); err != nil {
			log.WithError(err).Error("failed to remove existing standards library")
			return err
		}
	}

	if err := cloner.Clone(repository); err != nil {
		log.WithError(err).Error("failed to clone standards repository")
		return err
	}

	treePath := filepath.Join(repository.ClonePath, TreeFileName)
	log.WithFields(log.Fields{
		"clone_path": repository.ClonePath,
		"tree_path":  treePath,
	}).Info("generating standards index")

	if err := GenerateStandardsTree(repository.ClonePath, treePath); err != nil {
		log.WithError(err).Error("failed to generate standards index")
		return err
	}

	log.Info("successfully synced standards library")
	return nil
}

// Index regenerates the standards tree from an existing standards repository
// at clonePath without re-cloning. The tree is written to
// clonePath + "/" + TreeFileName. The ctx parameter is accepted for symmetry
// with the CLI command surface but is not currently propagated to filesystem
// operations.
func Index(ctx context.Context, clonePath string) error {
	treePath := filepath.Join(clonePath, TreeFileName)
	log.WithFields(log.Fields{
		"clone_path": clonePath,
		"tree_path":  treePath,
	}).Info("generating standards index")

	if err := GenerateStandardsTree(clonePath, treePath); err != nil {
		log.WithError(err).Error("failed to generate standards index")
		return err
	}

	log.Info("successfully generated standards index")
	return nil
}

func main() {
	cli := &cli.Command{
		Name:  "std-index",
		Usage: "Index and manage standards libraries",
		Commands: []*cli.Command{
			{
				Name:  "sync",
				Usage: "Sync and index a standards library",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "repository",
						Aliases:  []string{"r"},
						Usage:    "Git repository URL to clone",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "branch",
						Aliases: []string{"b"},
						Usage:   "Branch to checkout",
					},
					&cli.StringFlag{
						Name:    "tag",
						Aliases: []string{"t"},
						Usage:   "Tag to checkout",
					},
					&cli.StringFlag{
						Name:    "clone-path",
						Aliases: []string{"p"},
						Usage:   "Directory to clone standards into",
						Value:   DefaultClonePath(),
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					branch := cmd.String("branch")
					tag := cmd.String("tag")

					if branch != "" && tag != "" {
						return errors.New("only one of --branch or --tag can be specified, not both")
					}

					clonePath := ExpandPath(cmd.String("clone-path"))
					repo := GitRepository{
						Repository: cmd.String("repository"),
						Branch:     branch,
						Tag:        tag,
						ClonePath:  clonePath,
					}
					cloner := &ExecGitCloner{}
					if err := Sync(ctx, cloner, repo); err != nil {
						return err
					}
					PrintSuggestedInstructions(filepath.Join(clonePath, TreeFileName))
					return nil
				},
			},
			{
				Name:  "index",
				Usage: "Index a standards library",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "clone-path",
						Aliases: []string{"p"},
						Usage:   "Directory containing the standards to index",
						Value:   DefaultClonePath(),
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					clonePath := ExpandPath(cmd.String("clone-path"))
					if err := Index(ctx, clonePath); err != nil {
						return err
					}
					PrintSuggestedInstructions(filepath.Join(clonePath, TreeFileName))
					return nil
				},
			},
		},
	}

	if err := cli.Run(context.Background(), os.Args); err != nil {
		log.WithError(err).Error("failed to run std-index")
		os.Exit(1)
	}
}
