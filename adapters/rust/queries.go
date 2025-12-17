package rust

// rustCommentQuery is the Tree-sitter query to extract comments.
// It matches both line comments (//) and block comments (/* */).
const rustCommentQuery = `
(line_comment) @comment
(block_comment) @comment
`
