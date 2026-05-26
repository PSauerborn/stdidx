package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectAliases(t *testing.T) {
	t.Run("empty - no files declare aliases", func(t *testing.T) {
		files := []StandardsFile{
			{Path: "a.md", Header: StandardsHeader{}},
			{Path: "b.md", Header: StandardsHeader{}},
		}
		assert.Equal(t, map[string]string{}, CollectAliases(files))
	})

	t.Run("merge - aliases from multiple files combine", func(t *testing.T) {
		files := []StandardsFile{
			{
				Path: "a.md",
				Header: StandardsHeader{
					Aliases: map[string]string{"js": "javascript", "ts": "typescript"},
				},
			},
			{
				Path: "b.md",
				Header: StandardsHeader{
					Aliases: map[string]string{"py": "python"},
				},
			},
		}
		expected := map[string]string{
			"js": "javascript",
			"ts": "typescript",
			"py": "python",
		}
		assert.Equal(t, expected, CollectAliases(files))
	})

	t.Run("conflict - first declaration wins", func(t *testing.T) {
		files := []StandardsFile{
			{
				Path: "a.md",
				Header: StandardsHeader{
					Aliases: map[string]string{"db": "postgresql"},
				},
			},
			{
				Path: "b.md",
				Header: StandardsHeader{
					Aliases: map[string]string{"db": "mysql"},
				},
			},
		}
		merged := CollectAliases(files)
		assert.Equal(t, "postgresql", merged["db"])
	})

	t.Run("redundant declaration - same alias same canonical is not a conflict", func(t *testing.T) {
		files := []StandardsFile{
			{
				Path: "a.md",
				Header: StandardsHeader{
					Aliases: map[string]string{"js": "javascript"},
				},
			},
			{
				Path: "b.md",
				Header: StandardsHeader{
					Aliases: map[string]string{"js": "javascript"},
				},
			},
		}
		assert.Equal(t, map[string]string{"js": "javascript"}, CollectAliases(files))
	})
}

func TestNormalizeTopics(t *testing.T) {
	t.Run("noop - no aliases", func(t *testing.T) {
		files := []StandardsFile{
			{
				Path:   "a.md",
				Header: StandardsHeader{Topics: []string{"js", "ts"}},
			},
		}
		NormalizeTopics(files, map[string]string{})
		assert.Equal(t, []string{"js", "ts"}, files[0].Header.Topics)
	})

	t.Run("rewrite - aliased topics replaced with canonical", func(t *testing.T) {
		files := []StandardsFile{
			{
				Path:   "a.md",
				Header: StandardsHeader{Topics: []string{"js", "ts", "vue"}},
			},
		}
		aliases := map[string]string{"js": "javascript", "ts": "typescript"}
		NormalizeTopics(files, aliases)
		assert.Equal(t, []string{"javascript", "typescript", "vue"}, files[0].Header.Topics)
	})

	t.Run("dedupe - aliases that collapse onto an existing topic drop the duplicate", func(t *testing.T) {
		files := []StandardsFile{
			{
				Path:   "a.md",
				Header: StandardsHeader{Topics: []string{"javascript", "js", "vue"}},
			},
		}
		aliases := map[string]string{"js": "javascript"}
		NormalizeTopics(files, aliases)
		// `js` rewrites to `javascript`, which is already present at index 0 → dropped.
		assert.Equal(t, []string{"javascript", "vue"}, files[0].Header.Topics)
	})

	t.Run("dedupe - preserves first-occurrence order", func(t *testing.T) {
		files := []StandardsFile{
			{
				Path:   "a.md",
				Header: StandardsHeader{Topics: []string{"ts", "javascript", "js"}},
			},
		}
		aliases := map[string]string{"js": "javascript", "ts": "typescript"}
		NormalizeTopics(files, aliases)
		// `ts` → `typescript` first, then `javascript` (already canonical) preserved,
		// then `js` → `javascript` which is already seen → dropped.
		assert.Equal(t, []string{"typescript", "javascript"}, files[0].Header.Topics)
	})

	t.Run("non-transitive - chains do not collapse beyond one level", func(t *testing.T) {
		files := []StandardsFile{
			{
				Path:   "a.md",
				Header: StandardsHeader{Topics: []string{"a"}},
			},
		}
		// a → b → c is NOT a chain; `a` resolves to `b` only.
		aliases := map[string]string{"a": "b", "b": "c"}
		NormalizeTopics(files, aliases)
		assert.Equal(t, []string{"b"}, files[0].Header.Topics)
	})

	t.Run("multi-file - applies to every file's topics", func(t *testing.T) {
		files := []StandardsFile{
			{Path: "a.md", Header: StandardsHeader{Topics: []string{"js"}}},
			{Path: "b.md", Header: StandardsHeader{Topics: []string{"py"}}},
		}
		aliases := map[string]string{"js": "javascript", "py": "python"}
		NormalizeTopics(files, aliases)
		assert.Equal(t, []string{"javascript"}, files[0].Header.Topics)
		assert.Equal(t, []string{"python"}, files[1].Header.Topics)
	})
}
