package main

const (
	DefaultCloneDirName = ".stdidx"
	TreeFileName        = "standards-tree.yaml"
)

// SuggestedAgentInstructionsTemplate is a printf-style template for the prompt
// printed to the user after a successful sync or index. The single %s placeholder
// is substituted with the absolute path to the generated standards tree.
const SuggestedAgentInstructionsTemplate = `When working on a task, consult the standards tree in %s
to find applicable coding standards.

1. Always start at the root nodes. Read any root node whose scope
   matches the files you're working with or whose scope is "*".

2. For each node you read, check its children. Descend into a child
   if its scope or tags match your current context.

3. Stop descending a branch when no children match your context.

4. Collect all matching nodes from root to leaf. Standards at every
   level in the path apply — a child does not replace its parent,
   it adds to it.

5. If a child standard contradicts a parent, the child takes precedence.

To determine if a node matches your context:
- description: compare the description of the node to the task you're working on
- scope: compare against the file extensions you're editing
  ("*.py", "*.ts", "*" matches everything)
- topics: compare against the project's detected frameworks/tools
  (e.g. if package.json has "react" as a dependency, the "react"
  topic matches)`
