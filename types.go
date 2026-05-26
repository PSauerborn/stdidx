package main

// Scope is a list of glob patterns describing the file extensions a standard
// applies to. It unmarshals from either a single YAML scalar
// (`scope: '*.go'`) or a YAML sequence (`scope: ['*.go', '*.tsx']`), and is
// marshaled as a sequence on output. The scalar form is preserved for
// backward compatibility with standards files written before multi-extension
// scope was supported.
type Scope []string

// UnmarshalYAML accepts either a single string or a list of strings.
func (s *Scope) UnmarshalYAML(unmarshal func(any) error) error {
	var single string
	if err := unmarshal(&single); err == nil {
		*s = []string{single}
		return nil
	}
	var list []string
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = list
	return nil
}

// StandardsHeader is the YAML frontmatter of a standards Markdown file.
// Parent, when set, is a path to another standards file (relative to the
// standards repo root) and is used to build the hierarchical tree.
type StandardsHeader struct {
	Title       string   `yaml:"title" validate:"required"`
	Description string   `yaml:"description" validate:"required"`
	Scope       Scope    `yaml:"scope" validate:"required"`
	Topics      []string `yaml:"topics" validate:"required"`
	Parent      *string  `yaml:"parent" validate:"omitempty"`
}

// StandardsFile pairs a parsed StandardsHeader with the path of the file it
// was parsed from. Path is stored relative to the standards repo root and
// uses forward slashes regardless of host OS.
type StandardsFile struct {
	Path   string
	Header StandardsHeader
}

// Node is a single entry in the standards tree. Children holds nodes whose
// Parent field points at this node's Path. Parent is excluded from YAML
// output because the tree's hierarchy is encoded via nested Children, not
// back-pointers. Children is omitted from YAML output when empty.
type Node struct {
	Path        string
	Title       string
	Description string
	Scope       Scope
	Topics      []string
	Parent      *string `yaml:"-"`
	Children    []*Node `yaml:"children,omitempty"`
}

// StandardsTree is the top-level container serialized to standards-tree.yaml.
// GeneratedAt is an ISO 8601 UTC timestamp stamped at marshal time, useful
// for detecting stale indices. Nodes contains only root nodes (those without
// a Parent); descendant nodes are reached via Node.Children.
type StandardsTree struct {
	GeneratedAt string  `yaml:"generated_at,omitempty"`
	Nodes       []*Node `yaml:"nodes"`
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
