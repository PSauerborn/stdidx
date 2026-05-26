package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// getNodeCount returns the number of nodes in the tree
// including all children by recursively traversing the tree
func getNodeCount(nodes []*Node, count int) int {
	for _, node := range nodes {
		count++
		count = getNodeCount(node.Children, count)
	}
	return count
}

func init() {
	log.SetLevel(log.DebugLevel)
	// Pin the clock so GenerateStandardsTree's generated_at field is stable
	// across runs, enabling byte-equality golden-file comparison in tests.
	nowFn = func() time.Time {
		return time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC)
	}
}

func TestSync(t *testing.T) {
	t.Run("success - existing directory", func(t *testing.T) {
		ctx := context.Background()

		clonePath := filepath.Join("tests", "tmp")
		treePath := filepath.Join(clonePath, "standards-tree.yaml")

		// create temporary directory to use
		err := os.Mkdir(clonePath, 0755)
		assert.NoError(t, err)

		defer func() {
			if err := os.RemoveAll(clonePath); err != nil {
				t.Errorf("os.RemoveAll() error = %v", err)
			}
		}()

		// assert that treePath does not exists
		assert.NoFileExists(t, treePath)

		cloner := NewMockGitCloner("tests/mock_repository")
		repo := GitRepository{
			Repository: "https://github.com/golang/go",
			Branch:     "master",
			ClonePath:  clonePath,
		}

		err = Sync(ctx, cloner, repo)
		assert.NoError(t, err)

		// assert that treePath exists
		assert.FileExists(t, treePath)

		expected, err := os.ReadFile("tests/fixtures/expected_tree.yaml")
		if err != nil {
			t.Fatal(err)
		}

		actual, err := os.ReadFile(treePath)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, string(expected), string(actual))
	})

	t.Run("success - non-existing directory", func(t *testing.T) {
		ctx := context.Background()

		clonePath := filepath.Join("tests", "tmp")
		treePath := filepath.Join(clonePath, "standards-tree.yaml")

		defer func() {
			if err := os.RemoveAll(clonePath); err != nil {
				t.Errorf("os.RemoveAll() error = %v", err)
			}
		}()

		// assert that treePath does not exists
		assert.NoFileExists(t, treePath)

		cloner := NewMockGitCloner("tests/mock_repository")
		repo := GitRepository{
			Repository: "https://github.com/golang/go",
			Branch:     "master",
			ClonePath:  clonePath,
		}

		err := Sync(ctx, cloner, repo)
		assert.NoError(t, err)

		// assert that treePath exists
		assert.FileExists(t, treePath)

		expected, err := os.ReadFile("tests/fixtures/expected_tree.yaml")
		if err != nil {
			t.Fatal(err)
		}

		actual, err := os.ReadFile(treePath)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, string(expected), string(actual))
	})
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, info.Mode())
	})
}

func TestIndex(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		clonePath := filepath.Join("tests", "tmp")
		treePath := filepath.Join(clonePath, "standards-tree.yaml")

		err := os.Mkdir(clonePath, 0755)
		assert.NoError(t, err)

		defer func() {
			if err := os.RemoveAll(clonePath); err != nil {
				t.Errorf("os.RemoveAll() error = %v", err)
			}
		}()

		err = copyDir("tests/mock_repository", clonePath)
		assert.NoError(t, err)

		// assert that treePath does not exists
		assert.NoFileExists(t, treePath)

		err = Index(ctx, clonePath)
		assert.NoError(t, err)

		// assert that treePath exists
		assert.FileExists(t, treePath)

		expected, err := os.ReadFile("tests/fixtures/expected_tree.yaml")
		if err != nil {
			t.Fatal(err)
		}

		actual, err := os.ReadFile(treePath)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, string(expected), string(actual))
	})
}
