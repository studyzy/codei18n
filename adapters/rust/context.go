package rust

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// CommentType represents the type of Rust comment.
type CommentType int

const (
	// NormalComment represents standard comments (// or /* */).
	NormalComment CommentType = iota
	// DocComment represents outer documentation comments (/// or /** */).
	DocComment
	// ModuleDocComment represents inner/module documentation comments (//! or /*! */).
	ModuleDocComment
)

// IdentifyCommentType determines the type of comment based on its content prefix.
func IdentifyCommentType(content string) CommentType {
	if strings.HasPrefix(content, "///") || strings.HasPrefix(content, "/**") {
		return DocComment
	}
	if strings.HasPrefix(content, "//!") || strings.HasPrefix(content, "/*!") {
		return ModuleDocComment
	}
	return NormalComment
}

// FindOwnerNode finds the semantic owner node of the comment.
// For DocComments, it looks for the next semantic sibling (skipping attributes).
// For ModuleDocComments and NormalComments, it returns the parent node (representing the scope).
func FindOwnerNode(commentNode *sitter.Node, src []byte) *sitter.Node {
	content := commentNode.Content(src)
	cType := IdentifyCommentType(content)

	if cType == DocComment {
		curr := commentNode.NextNamedSibling()
		for curr != nil {
			// Skip attributes (#[...]) to find the actual item
			if curr.Type() == "attribute_item" {
				curr = curr.NextNamedSibling()
				continue
			}
			// Skip other comments to find the actual item they are documenting
			if curr.Type() == "line_comment" || curr.Type() == "block_comment" {
				curr = curr.NextNamedSibling()
				continue
			}
			return curr
		}
		// Fallback to parent if no sibling found (orphaned doc comment)
		return commentNode.Parent()
	}

	// ModuleDoc and NormalDoc belong to Parent scope
	return commentNode.Parent()
}
