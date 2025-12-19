package java

// javaCommentQuery 定义用于提取 Java 注释的 Tree-sitter 查询
// 匹配行注释 (//)、块注释 (/* */) 和 Javadoc (/** */)
const javaCommentQuery = `
(line_comment) @comment
(block_comment) @comment
`
