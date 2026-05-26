package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// nowFn returns the current time. Exposed as a package variable so tests can
// pin the clock for deterministic golden-file comparisons.
var nowFn = time.Now

// ExtractMDHeader reads the Markdown file at path and parses its YAML
// frontmatter into a StandardsHeader. A file whose frontmatter fails struct
// validation (e.g. missing required fields, or no frontmatter at all) returns
// (nil, nil) rather than an error — this lets callers silently skip arbitrary
// Markdown files (READMEs, etc.) that happen to live in a standards repo.
// Returns a non-nil error only when the file cannot be read or its frontmatter
// is syntactically malformed.
func ExtractMDHeader(path string) (*StandardsHeader, error) {
	log.WithFields(log.Fields{
		"path": path,
	}).Debug("extracting md header")

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var header StandardsHeader
	reader := strings.NewReader(string(content))
	if _, err := frontmatter.Parse(reader, &header); err != nil {
		return nil, err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(header); err != nil {
		log.WithError(err).Debug("failed to validate md header")
		return nil, nil
	}
	return &header, nil
}

// ParseMDDocuments recursively walks root and returns a StandardsFile for
// every Markdown file with valid frontmatter. Files without valid frontmatter
// are skipped with a warning. Each StandardsFile.Path is stored as a slash-
// separated path relative to root, matching the form `parent:` is authored
// in. BuildHierarchy can therefore look parents up directly without further
// rewriting.
func ParseMDDocuments(root string) ([]StandardsFile, error) {
	headers := make([]StandardsFile, 0)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		header, err := ExtractMDHeader(path)
		if err != nil {
			return err
		}

		if header == nil {
			log.WithFields(log.Fields{
				"path": path,
			}).Warn("found markdown file without valid header. skipping.")
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		headers = append(headers, StandardsFile{
			Path:   filepath.ToSlash(relPath),
			Header: *header,
		})
		return nil
	})

	return headers, err
}

// BuildHierarchy assembles a StandardsTree from a flat list of StandardsFile.
// Files without a Parent become root nodes. Files with a Parent are attached
// as Children of the node whose Path equals the parent value; files whose
// parent does not resolve are logged and dropped. Children at every level
// are sorted alphabetically by Title.
func BuildHierarchy(files []StandardsFile) StandardsTree {

	nodes := map[string]*Node{}

	for _, file := range files {
		nodes[file.Path] = &Node{
			Title:       file.Header.Title,
			Description: file.Header.Description,
			Parent:      file.Header.Parent,
			Scope:       file.Header.Scope,
			Topics:      file.Header.Topics,
			Path:        file.Path,
		}
	}

	roots := []*Node{}

	for _, file := range files {
		if file.Header.Parent == nil {
			node := nodes[file.Path]
			roots = append(roots, node)
		} else {
			parent, exists := nodes[*file.Header.Parent]
			if !exists {
				log.WithFields(log.Fields{
					"path":   file.Path,
					"parent": *file.Header.Parent,
				}).Warn("found node with parent that does not exist. skipping.")
				continue
			}
			node := nodes[file.Path]
			parent.Children = append(parent.Children, node)
		}
	}

	sortChildren(roots)
	return StandardsTree{Nodes: roots}
}

// sortChildren sorts the provided slice alphabetically by Title in place and
// recurses into each node's Children. The resulting deterministic order is
// what makes byte-for-byte golden-file comparisons in tests possible.
func sortChildren(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Title < nodes[j].Title
	})
	for _, node := range nodes {
		sortChildren(node.Children)
	}
}

// CollectAliases merges the `aliases:` declarations from every parsed file
// into a single global map. When two files declare the same alias key with
// different canonical values, the first one encountered (in walk order)
// wins and the conflict is logged at warn level. Returns an empty map when
// no file declares any aliases.
func CollectAliases(files []StandardsFile) map[string]string {
	merged := map[string]string{}
	for _, file := range files {
		for alias, canonical := range file.Header.Aliases {
			if existing, ok := merged[alias]; ok && existing != canonical {
				log.WithFields(log.Fields{
					"alias":     alias,
					"canonical": existing,
					"ignored":   canonical,
					"source":    file.Path,
				}).Warn("conflicting alias declaration; keeping first")
				continue
			}
			merged[alias] = canonical
		}
	}
	return merged
}

// NormalizeTopics rewrites each file's Topics in place: any topic whose
// value is a key in aliases is replaced by its canonical form, and the
// resulting list is deduplicated while preserving first-occurrence order.
// Aliasing is non-transitive (single-level): if `a → b` and `b → c` are
// both declared, `a` resolves to `b`. No-op when aliases is empty.
func NormalizeTopics(files []StandardsFile, aliases map[string]string) {
	if len(aliases) == 0 {
		return
	}
	for i, file := range files {
		seen := map[string]bool{}
		out := make([]string, 0, len(file.Header.Topics))
		for _, t := range file.Header.Topics {
			if canonical, ok := aliases[t]; ok {
				t = canonical
			}
			if seen[t] {
				continue
			}
			seen[t] = true
			out = append(out, t)
		}
		files[i].Header.Topics = out
	}
}

// GenerateStandardsTree walks the standards repository rooted at clonePath,
// builds the standards tree from each Markdown file's frontmatter, applies
// any declared topic aliases, stamps the current time as GeneratedAt, and
// writes the YAML-serialized tree to outputPath. The parent directory of
// outputPath must already exist.
func GenerateStandardsTree(clonePath, outputPath string) error {
	log.WithFields(log.Fields{
		"clone_path":  clonePath,
		"output_path": outputPath,
	}).Debug("parsing standards files")

	headers, err := ParseMDDocuments(clonePath)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"count": len(headers),
	}).Debug("creating standards tree")

	aliases := CollectAliases(headers)
	NormalizeTopics(headers, aliases)

	tree := BuildHierarchy(headers)
	tree.GeneratedAt = nowFn().UTC().Format(time.RFC3339)

	data, err := yaml.Marshal(tree)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return err
	}
	return nil
}
