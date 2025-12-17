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

			// Calculate Range (Trimming trailing newline if present for Line/Doc comments)
			startRow := int(node.StartPoint().Row) + 1
			startCol := int(node.StartPoint().Column) + 1
			endRow := int(node.EndPoint().Row) + 1
			endCol := int(node.EndPoint().Column) + 1
			
			// Fix for Tree-sitter including newline in Line/Doc comments
			// If it's a line/doc comment and includes newline (EndCol == 1 && EndRow > StartRow)
			// We want to snap it back to the same line.
			cType := getDomainCommentType(content)
			if (cType == domain.CommentTypeLine || cType == domain.CommentTypeDoc) && 
			   strings.HasSuffix(content, "\n") {
				// Trim the newline from content
				content = strings.TrimSuffix(content, "\n")
				// Recalculate EndRow/EndCol based on trimmed content
				// Since it's a single line comment, EndRow is StartRow
				endRow = startRow
				// EndCol is StartCol + length (in bytes)
				endCol = startCol + len(content)
			}

			comment := &domain.Comment{
				SourceText: content,
				Symbol:     symbolPath,
				File:       file,
				Language:   "rust",
				Range: domain.TextRange{
					StartLine: startRow,
					StartCol:  startCol,
					EndLine:   endRow,
					EndCol:    endCol,
				},
				Type: cType,
			}
			comment.ID = utils.GenerateCommentID(comment)
			comments = append(comments, comment)
		}
	}

	return mergeComments(comments), nil
}

// mergeComments merges consecutive comments of the same type and symbol.
func mergeComments(comments []*domain.Comment) []*domain.Comment {
	if len(comments) == 0 {
		return comments
	}

	var merged []*domain.Comment
	for _, current := range comments {
		if len(merged) == 0 {
			merged = append(merged, current)
			continue
		}

		last := merged[len(merged)-1]

		// Condition to merge:
		// 1. Same Type (Doc or Line)
		// 2. Same Symbol
		// 3. Consecutive lines:
		//    - Since we fixed range to exclude newline, EndLine of last + 1 == StartLine of current
		isConsecutive := last.Range.EndLine+1 == current.Range.StartLine

		shouldMerge := (last.Type == domain.CommentTypeDoc || last.Type == domain.CommentTypeLine) &&
			last.Type == current.Type &&
			last.Symbol == current.Symbol &&
			isConsecutive

		if shouldMerge {
			// Append content
			// Ensure we have a newline between them if not present in the last one
			if !strings.HasSuffix(last.SourceText, "\n") {
				last.SourceText += "\n"
			}
			last.SourceText += current.SourceText

			// Update range
			last.Range.EndLine = current.Range.EndLine
			last.Range.EndCol = current.Range.EndCol

			// Regenerate ID because content changed
			last.ID = utils.GenerateCommentID(last)
		} else {
			merged = append(merged, current)
		}
	}
	return merged
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
