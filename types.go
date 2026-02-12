package main

type StandardsHeader struct {
	Title       string   `yaml:"title" validate:"required"`
	Description string   `yaml:"description" validate:"required"`
	Scope       string   `yaml:"scope" validate:"required"`
	Topics      []string `yaml:"topics" validate:"required"`
	Parent      *string  `yaml:"parent" validate:"omitempty"`
}

type StandardsFile struct {
	Path   string
	Header StandardsHeader
}

type Node struct {
	Path        string
	Title       string
	Description string
	Scope       string
	Topics      []string
	Parent      *string `yaml:"-"`
	Children    []*Node
}

type StandardsTree struct {
	Nodes []*Node
}

type GitRepository struct {
	Repository string `yaml:"repository" validate:"required"`
	Branch     string `yaml:"branch" validate:"omitempty"`
	Tag        string `yaml:"tag" validate:"omitempty"`
	ClonePath  string `yaml:"clone_path" validate:"required"`
}
