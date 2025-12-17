package adapters

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/studyzy/codei18n/adapters/golang"
	"github.com/studyzy/codei18n/adapters/rust"
	"github.com/studyzy/codei18n/core"
)

// GetAdapter returns the appropriate LanguageAdapter for the given file
func GetAdapter(filename string) (core.LanguageAdapter, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return golang.NewAdapter(), nil
	case ".rs":
		return rust.NewRustAdapter(), nil
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}
