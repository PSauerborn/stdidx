package main

// StandardsHeader is the YAML frontmatter of a standards Markdown file.
// Parent, when set, is a path to another standards file and is used to build
// the hierarchical tree.
type StandardsHeader struct {
	Title       string   `yaml:"title" validate:"required"`
	Description string   `yaml:"description" validate:"required"`
	Scope       string   `yaml:"scope" validate:"required"`
	Topics      []string `yaml:"topics" validate:"required"`
	Parent      *string  `yaml:"parent" validate:"omitempty"`
}

// StandardsFile pairs a parsed StandardsHeader with the path of the file it
// was parsed from. Path is the path returned by filepath.WalkDir relative to
// the directory the walk was rooted at.
type StandardsFile struct {
	Path   string
	Header StandardsHeader
}

// Node is a single entry in the standards tree. Children holds nodes whose
// Parent field points at this node's Path. Parent is excluded from YAML output
// because the tree's hierarchy is encoded via nested Children, not back-pointers.
type Node struct {
	Path        string
	Title       string
	Description string
	Scope       string
	Topics      []string
	Parent      *string `yaml:"-"`
	Children    []*Node
}

// StandardsTree is the top-level container serialized to standards-tree.yaml.
// Nodes contains only root nodes (those without a Parent); descendant nodes
// are reached via Node.Children.
type StandardsTree struct {
	Nodes []*Node
}

// GitRepository describes the remote standards repository to sync and the
// local directory it should be cloned into. Exactly one of Branch or Tag may
// be set; setting both is rejected by the CLI.
type GitRepository struct {
	Repository string `yaml:"repository" validate:"required"`
	Branch     string `yaml:"branch" validate:"omitempty"`
	Tag        string `yaml:"tag" validate:"omitempty"`
	ClonePath  string `yaml:"clone_path" validate:"required"`
}
