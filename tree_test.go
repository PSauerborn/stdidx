package main

import (
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestExtractMDHeader(t *testing.T) {
	t.Run("valid header", func(t *testing.T) {
		header, err := ExtractMDHeader("tests/mock_repository/golang/API.md")
		assert.NoError(t, err)
		assert.NotNil(t, header)

		assert.Equal(t, "Golang REST API Standards", header.Title)
		assert.Equal(t, "Standards for writing REST APIs in Go.", header.Description)
		assert.Equal(t, "golang/GENERAL.md", *header.Parent)
		assert.Equal(t, []string{"golang", "api", "rest", "gin-gonic"}, header.Topics)
		assert.Equal(t, "*.go", header.Scope)
	})

	t.Run("invalid header", func(t *testing.T) {
		header, err := ExtractMDHeader("tests/mock_repository/INVALID.md")
		assert.NoError(t, err)
		assert.Nil(t, header)
	})

	t.Run("no header", func(t *testing.T) {
		header, err := ExtractMDHeader("tests/mock_repository/README.md")
		assert.NoError(t, err)
		assert.Nil(t, header)
	})
}

func TestParseMDDocuments(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		headers, err := ParseMDDocuments("tests/mock_repository")
		assert.NoError(t, err)
		assert.NotNil(t, headers)

		assert.Equal(t, 5, len(headers))

		for _, h := range headers {
			assert.NotNil(t, h)

			if h.Header.Parent == nil {
				continue
			}
			parent := *h.Header.Parent
			if parent != "" {
				// all parent paths should be relative to the root directory,
				// NOT relative to the "cloned" repository
				assert.True(t, strings.HasPrefix(parent, "tests/mock_repository"))
			}
		}
	})
}

func TestBuildHierarchy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		files := []StandardsFile{
			{
				Path: "tests/tmp/golang/frameworks/GIN-GONIC.md",
				Header: StandardsHeader{
					Title:       "Golang Gin-Gonic Standards",
					Description: "Standards for using Gin Gonic framework.",
					Topics:      []string{"golang", "api", "gin-gonic"},
					Scope:       "*.go",
				},
			},
			{
				Path: "tests/tmp/golang/API.md",
				Header: StandardsHeader{
					Title:       "Golang REST API Standards",
					Description: "Standards for writing REST APIs in Go.",
					Topics:      []string{"golang", "api", "rest", "gin-gonic"},
					Scope:       "*.go",
				},
			},
			{
				Path: "tests/tmp/golang/WORKER.md",
				Header: StandardsHeader{
					Title:       "Golang Worker Standards",
					Description: "Standards for writing background workers in Go.",
					Topics:      []string{"golang", "worker", "rabbitmq", "message-broker", "amqp"},
					Scope:       "*.go",
				},
			},
			{
				Path: "tests/tmp/golang/GENERAL.md",
				Header: StandardsHeader{
					Title:       "Golang General Standards",
					Description: "General standards for writing Go applications.",
					Topics:      []string{"golang"},
					Scope:       "*.go",
				},
			},
			{
				Path: "tests/tmp/GENERAL.md",
				Header: StandardsHeader{
					Title:       "General Code Standards",
					Description: "Cross-language general coding standards and best practices.",
					Topics:      []string{"general", "docker", "makefiles", "pre-commit"},
					Scope:       "*",
				},
			},
		}

		files[0].Header.Parent = &files[1].Path
		files[1].Header.Parent = &files[3].Path
		files[2].Header.Parent = &files[3].Path

		tree := BuildHierarchy(files)
		assert.Equal(t, 5, getNodeCount(tree.Nodes, 0))

		// count root nodes
		assert.Equal(t, 2, len(tree.Nodes))

		content, err := os.ReadFile("tests/fixtures/expected_tree.yaml")
		if err != nil {
			t.Fatal(err)
		}

		encoded, err := yaml.Marshal(tree)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, string(encoded), string(content))
	})
}
