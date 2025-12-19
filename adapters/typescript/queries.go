package typescript

import (
	_ "embed"
)

// We can define queries here.
// For now we'll put the query string in a variable, but in production we might use embed.

// queryTS is a tree-sitter query to find comments in TypeScript/JavaScript
// Note: We want to capture comments. In tree-sitter, comments are often extra nodes
// or need specific queries if they are not part of the named grammar but are "extras".
// However, most tree-sitter parsers expose (comment) nodes.

const queryTS = `
(comment) @comment
`

// In some grammars (like TS/JS), comments are just (comment) nodes.
// We might need to distinguish between block and line comments if the grammar supports it.
// Usually:
// // ... is a comment
// /* ... */ is a comment
//
// We will simply iterate over all captures of (comment).
