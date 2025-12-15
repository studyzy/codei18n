package domain

// TextRange represents the range of text in the source code (1-based)
type TextRange struct {
	StartLine int `json:"startLine"` // Start line number
	StartCol  int `json:"startCol"`  // Start column number
	EndLine   int `json:"endLine"`   // End line number
	EndCol    int `json:"endCol"`    // End column number
}

// CommentType defines the type of comment
type CommentType string

const (
	CommentTypeLine  CommentType = "line"  // Single line comment //
	CommentTypeBlock CommentType = "block" // Block comment /* */
	CommentTypeDoc   CommentType = "doc"   // Documentation comment
)

// Comment represents a single comment extracted from AST
type Comment struct {
	// ID is the unique identifier calculated based on semantics
	// Rule: SHA1(file_path + language + symbol + normalized_text)
	ID string `json:"id"`

	// File is the file path relative to project root
	File string `json:"file"`

	// Language is the programming language identifier (e.g., "go", "rust")
	Language string `json:"language"`

	// Symbol is the semantic symbol path (e.g., "package.main.CalculateBalance")
	// For comments not bound to specific symbols, use "file.global" or similar
	Symbol string `json:"symbol"`

	// Range is the position of the comment in the source code
	Range TextRange `json:"range"`

	// SourceText is the original text of the comment (English)
	SourceText string `json:"sourceText"`

	// Type is the type of the comment
	Type CommentType `json:"type"`

	// LocalizedText is the translated text (Optional, populated during scan --with-translations)
	LocalizedText string `json:"localizedText,omitempty"`
}

// LocalizedComment represents the localized version of a comment
type LocalizedComment struct {
	// CommentID references the corresponding Comment.ID
	CommentID string `json:"commentId"`

	// Lang is the target language code (BCP 47, e.g., "zh-CN")
	Lang string `json:"lang"`

	// Text is the translated text
	Text string `json:"text"`
}

// Mapping stores the multi-language mappings for the entire project
type Mapping struct {
	// Version is the version of the mapping file format (e.g., "1.0")
	Version string `json:"version"`

	// SourceLanguage is the language in source code (usually "en")
	SourceLanguage string `json:"sourceLanguage"`

	// TargetLanguage is the local target language (e.g., "zh-CN")
	TargetLanguage string `json:"targetLanguage"`

	// Comments stores the mapping data
	// First Level Key: Comment.ID
	// Second Level Key: Language Code (e.g., "zh-CN")
	// Value: Translated Text
	Comments map[string]map[string]string `json:"comments"`
}
