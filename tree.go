package main

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

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

		// read contents of file and parse frontmatter.
		header, err := ExtractMDHeader(path)
		if err != nil {
			return err
		}

		if header != nil {
			headers = append(headers, StandardsFile{
				Path:   path,
				Header: *header,
			})
		} else {
			log.WithFields(log.Fields{
				"path": path,
			}).Warn("found markdown file without valid header. skipping.")
		}

		return nil
	})

	// ensure that the parent path is absolute
	// this is required as relationships are defined
	// relative to the directory the code is cloned into,
	// but the tree is built from the root of the repository.
	for i, file := range headers {
		if file.Header.Parent != nil {
			augmentedPath := path.Join(root, *file.Header.Parent)
			headers[i].Header.Parent = &augmentedPath
		}
	}
	return headers, err
}

// BuildHierarchy builds a nested tree from a flat list of headers. Headers
// without a Parent are root nodes. Headers with a Parent are nested under the
// node whose Scope matches the parent value.
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
			Children:    []*Node{},
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

func sortChildren(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Title < nodes[j].Title
	})
	for _, node := range nodes {
		sortChildren(node.Children)
	}
}

func GenerateStandardsTree(path string) error {
	log.WithFields(log.Fields{
		"path": path,
	}).Debug("parsing standards files")

	headers, err := ParseMDDocuments(path)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"count": len(headers),
	}).Debug("creating standards tree")

	tree := BuildHierarchy(headers)

	data, err := yaml.Marshal(tree)
	if err != nil {
		return err
	}

	outputPath := "standards-tree.yaml"
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return err
	}
	return nil
}
