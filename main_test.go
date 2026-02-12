package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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
}

func TestSync(t *testing.T) {
	t.Run("success - existing directory", func(t *testing.T) {
		ctx := context.Background()

		treePath := "standards-tree.yaml"
		// assert that treePath does not exists
		assert.NoFileExists(t, treePath)

		// create temporary directory to use
		err := os.Mkdir("tests/tmp", 0755)
		assert.NoError(t, err)

		defer func() {
			if err := os.RemoveAll("tests/tmp"); err != nil {
				t.Errorf("os.RemoveAll() error = %v", err)
			}
		}()

		cloner := NewMockGitCloner("tests/mock_repository")
		repo := GitRepository{
			Repository: "https://github.com/golang/go",
			Branch:     "master",
			ClonePath:  "tests/tmp",
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

		err = os.Remove(treePath)
		assert.NoError(t, err)
	})

	t.Run("success - non-existing directory", func(t *testing.T) {
		ctx := context.Background()

		treePath := "standards-tree.yaml"
		// assert that treePath does not exists
		assert.NoFileExists(t, treePath)

		clonePath := filepath.Join("tests", "tmp")
		defer func() {
			if err := os.RemoveAll(clonePath); err != nil {
				t.Errorf("os.RemoveAll() error = %v", err)
			}
		}()

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

		err = os.Remove(treePath)
		assert.NoError(t, err)
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

		treePath := "standards-tree.yaml"
		// assert that treePath does not exists
		assert.NoFileExists(t, treePath)

		err := os.Mkdir("tests/tmp", 0755)
		assert.NoError(t, err)

		err = copyDir("tests/mock_repository", "tests/tmp")
		assert.NoError(t, err)

		path := filepath.Join("tests", "tmp")
		err = Index(ctx, path)
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

		err = os.Remove(treePath)
		if err != nil {
			t.Fatal(err)
		}

		err = os.RemoveAll("tests/tmp")
		if err != nil {
			t.Fatal(err)
		}
	})
}
