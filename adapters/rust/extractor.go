package rust

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/studyzy/codei18n/core/domain"
	"github.com/studyzy/codei18n/core/utils"
)

// extractComments processes the AST and extracts all comments matching the query.
func (a *RustAdapter) extractComments(root *sitter.Node, src []byte, file string) ([]*domain.Comment, error) {
	q, err := sitter.NewQuery([]byte(rustCommentQuery), rust.GetLanguage())
	if err != nil {
		return nil, err
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(q, root)
	defer qc.Close() // Ensure cursor is closed

	var comments []*domain.Comment

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, c := range m.Captures {
			node := c.Node
			content := node.Content(src)

			if isEmptyComment(content) {
				continue
			}

			// Find Owner and Symbol Path
			owner := FindOwnerNode(node, src)
			symbolPath := ResolveSymbolPath(owner, src)

			comment := &domain.Comment{
				SourceText: content,
				Symbol:     symbolPath,
				File:       file,
				Language:   "rust",
				Range: domain.TextRange{
					StartLine: int(node.StartPoint().Row) + 1,
					StartCol:  int(node.StartPoint().Column) + 1,
					EndLine:   int(node.EndPoint().Row) + 1,
					EndCol:    int(node.EndPoint().Column) + 1,
				},
				Type: getDomainCommentType(content),
			}
			comment.ID = utils.GenerateCommentID(comment)
			comments = append(comments, comment)
		}
	}

	return comments, nil
}

// getDomainCommentType maps raw comment content to domain.CommentType
func getDomainCommentType(content string) domain.CommentType {
	if strings.HasPrefix(content, "/*") {
		return domain.CommentTypeBlock
	}
	if strings.HasPrefix(content, "///") || strings.HasPrefix(content, "//!") {
		return domain.CommentTypeDoc
	}
	return domain.CommentTypeLine
}

func isEmptyComment(content string) bool {
	trimmed := strings.TrimSpace(content)
	switch {
	case strings.HasPrefix(trimmed, "///"):
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "///")) == ""
	case strings.HasPrefix(trimmed, "//!"):
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "//!")) == ""
	case strings.HasPrefix(trimmed, "//"):
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "//")) == ""
	case strings.HasPrefix(trimmed, "/*"):
		inner := strings.TrimPrefix(trimmed, "/*")
		if strings.HasSuffix(inner, "*/") {
			inner = strings.TrimSuffix(inner, "*/")
		}
		return strings.TrimSpace(inner) == ""
	default:
		return strings.TrimSpace(trimmed) == ""
	}
}