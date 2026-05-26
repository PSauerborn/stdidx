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

1. Locate every node — at any depth in the tree — whose scope, topics, or
   description matches your current context. A node matches on:

   - scope: any entry is "*" or matches a file extension you're editing.
   - topics: any topic matches the project's detected frameworks/tools
     (e.g. if package.json declares "react" as a dependency, the "react"
     topic matches).
   - description: the description semantically describes the task you're
     working on.

2. For each matched node, also read every ancestor in its chain back to the
   root. Standards at every level on the path apply — a parent is not
   optional just because the parent itself wasn't directly matched.

3. If a child standard contradicts an ancestor, the child takes precedence.

4. The "path" field on each node is relative to the directory the standards
   were synced into. Prepend the clone path (default ~/.stdidx) when reading.`
