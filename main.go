package main

import (
	"context"
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

func PrintSuggestedInstructions() {
	println("\nDon't forget to instruct your agent to use the standards index. Suggested prompt:")
	println("\n" + SuggestedAgentInstructions + "\n")
}

func Sync(ctx context.Context, cloner GitCloner, repository GitRepository) error {
	log.WithFields(log.Fields{
		"repository": repository.Repository,
		"branch":     repository.Branch,
		"tag":        repository.Tag,
	}).Info("syncing standards library")

	// check if already exists
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

	log.WithFields(log.Fields{
		"clone_path": repository.ClonePath,
	}).Info("generating standards index")

	if err := GenerateStandardsTree(repository.ClonePath); err != nil {
		log.WithError(err).Error("failed to generate standards index")
		return err
	}

	log.Info("successfully synced standards library")
	return nil
}

func Index(ctx context.Context, clonePath string) error {
	log.WithFields(log.Fields{
		"clone_path": clonePath,
	}).Info("generating standards index")

	if err := GenerateStandardsTree(clonePath); err != nil {
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
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					branch := cmd.String("branch")
					tag := cmd.String("tag")

					if branch != "" && tag != "" {
						return errors.New("only one of --branch or --tag can be specified, not both")
					}

					repo := GitRepository{
						Repository: cmd.String("repository"),
						Branch:     branch,
						Tag:        tag,
						ClonePath:  DefaultClonePath,
					}
					cloner := &ExecGitCloner{}
					if err := Sync(ctx, cloner, repo); err != nil {
						return err
					}
					PrintSuggestedInstructions()
					return nil
				},
			},
			{
				Name:  "index",
				Usage: "Index a standards library",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if err := Index(ctx, DefaultClonePath); err != nil {
						return err
					}
					PrintSuggestedInstructions()
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
