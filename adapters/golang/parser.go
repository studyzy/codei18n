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
		for _, c := range cg.List {
			symbol := commentSymbols[c]
			if symbol == "" {
				symbol = "package." + f.Name.Name // Default to package level
			}
			comments = append(comments, a.createComment(fset, c, filePath, symbol))
		}
	}

	return comments, nil
}

func (a *Adapter) createComment(fset *token.FileSet, c *ast.Comment, file, symbol string) *domain.Comment {
	pos := fset.Position(c.Pos())
	end := fset.Position(c.End())

	// Clean the text: remove // or /* */
	text := c.Text
	// Determine type
	cType := domain.CommentTypeLine
	if strings.HasPrefix(text, "/*") {
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
		SourceText: text,
		Type:       cType,
	}
}
