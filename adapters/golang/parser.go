package golang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/studyzy/codei18n/core/domain"
)

// Adapter implements core.LanguageAdapter for Go
type Adapter struct{}

// NewAdapter creates a new Go language adapter
func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Language() string {
	return "go"
}

// Parse parses the Go source code and extracts comments
func (a *Adapter) Parse(file string, src []byte) ([]*domain.Comment, error) {
	fset := token.NewFileSet()

	// Parse the file
	// ParseComments is required to get comments in the AST
	f, err := parser.ParseFile(fset, file, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Use scope tracking to find comment symbols
	return a.parseWithScopeTracking(fset, f, file)
}

func (a *Adapter) parseWithScopeTracking(fset *token.FileSet, f *ast.File, filePath string) ([]*domain.Comment, error) {
	var comments []*domain.Comment

	// Let's use ast.NewCommentMap to find attached comments.
	cmap := ast.NewCommentMap(fset, f, f.Comments)

	// Helper to determine symbol name from a node
	getSymbol := func(n ast.Node) string {
		switch t := n.(type) {
		case *ast.File:
			return "package." + t.Name.Name
		case *ast.FuncDecl:
			return "func." + t.Name.Name
		case *ast.GenDecl:
			// import, const, type, var
			return "decl." + t.Tok.String()
		case *ast.TypeSpec:
			return "type." + t.Name.Name
		case *ast.ValueSpec:
			if len(t.Names) > 0 {
				return "var." + t.Names[0].Name
			}
		}
		return "unknown"
	}

	// We need to map *ast.Comment -> Symbol
	commentSymbols := make(map[*ast.Comment]string)

	for node, groups := range cmap {
		symbol := getSymbol(node)
		if symbol == "unknown" {
			continue // keep default or try parent
		}
		for _, cg := range groups {
			for _, c := range cg.List {
				// Only set if not already set (inner nodes might be more specific)
				// But CommentMap maps to the "best" node usually.
				commentSymbols[c] = symbol
			}
		}
	}

	// Now iterate in order
	for _, cg := range f.Comments {
		if len(cg.List) == 0 {
			continue
		}

		// Use the symbol of the first comment in the group
		symbol := commentSymbols[cg.List[0]]
		if symbol == "" {
			symbol = "package." + f.Name.Name // Default to package level
		}

		// Split group into subgroups and create comments
		subComments := a.processCommentGroup(fset, cg, filePath, symbol)
		comments = append(comments, subComments...)
	}

	return comments, nil
}

// processCommentGroup handles the logic of merging or splitting comments within a group.
// It ensures that only consecutive Line comments (//) are merged.
// Block comments (/* */) are kept separate, even if consecutive or adjacent to line comments.
func (a *Adapter) processCommentGroup(fset *token.FileSet, cg *ast.CommentGroup, file, symbol string) []*domain.Comment {
	var results []*domain.Comment
	if len(cg.List) == 0 {
		return results
	}

	// Buffer for consecutive line comments
	var lineBuffer []*ast.Comment

	flushLineBuffer := func() {
		if len(lineBuffer) == 0 {
			return
		}
		results = append(results, a.createMergedComment(fset, lineBuffer, file, symbol))
		lineBuffer = nil // Clear buffer
	}

	for _, c := range cg.List {
		if strings.HasPrefix(c.Text, "//") {
			// It's a line comment, add to buffer
			lineBuffer = append(lineBuffer, c)
		} else {
			// It's a block comment (/* ... */)
			// 1. Flush any existing line comments
			flushLineBuffer()

			// 2. Add this block comment individually
			results = append(results, a.createSingleComment(fset, c, file, symbol))
		}
	}
	// Flush remaining line comments at the end
	flushLineBuffer()

	return results
}

// createMergedComment creates a single comment from a list of Line comments
func (a *Adapter) createMergedComment(fset *token.FileSet, comments []*ast.Comment, file, symbol string) *domain.Comment {
	if len(comments) == 0 {
		return nil
	}

	first := comments[0]
	last := comments[len(comments)-1]

	pos := fset.Position(first.Pos())
	end := fset.Position(last.End())

	var sb strings.Builder
	for i, c := range comments {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(c.Text)
	}

	return &domain.Comment{
		File:     file,
		Language: "go",
		Symbol:   symbol,
		Range: domain.TextRange{
			StartLine: pos.Line,
			StartCol:  pos.Column,
			EndLine:   end.Line,
			EndCol:    end.Column,
		},
		SourceText: sb.String(),
		Type:       domain.CommentTypeLine,
	}
}

// createSingleComment creates a comment from a single AST comment (usually Block)
func (a *Adapter) createSingleComment(fset *token.FileSet, c *ast.Comment, file, symbol string) *domain.Comment {
	pos := fset.Position(c.Pos())
	end := fset.Position(c.End())

	cType := domain.CommentTypeLine
	if strings.HasPrefix(c.Text, "/*") {
		cType = domain.CommentTypeBlock
	}

	return &domain.Comment{
		File:     file,
		Language: "go",
		Symbol:   symbol,
		Range: domain.TextRange{
			StartLine: pos.Line,
			StartCol:  pos.Column,
			EndLine:   end.Line,
			EndCol:    end.Column,
		},
		SourceText: c.Text,
		Type:       cType,
	}
}
