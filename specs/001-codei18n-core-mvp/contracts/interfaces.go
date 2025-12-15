package contracts

// TextRange represents the position of text in the source file
type TextRange struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

// CommentType defines the type of comment
type CommentType string

const (
	CommentTypeLine  CommentType = "line"
	CommentTypeBlock CommentType = "block"
	CommentTypeDoc   CommentType = "doc"
)

// Comment represents a comment extracted from source code
type Comment struct {
	ID         string
	File       string
	Language   string
	Symbol     string
	Range      TextRange
	SourceText string
	Type       CommentType
}

// LanguageAdapter defines the interface for language-specific parsers
type LanguageAdapter interface {
	// Language returns the language identifier (e.g., "go")
	Language() string

	// Parse scans a file and returns all comments
	Parse(filePath string, content []byte) ([]Comment, error)
}

// Translator defines the interface for translation services
type Translator interface {
	// Translate translates a list of texts from source language to target language
	// Returns a map of original text to translated text
	Translate(texts []string, from, to string) (map[string]string, error)
}

// MappingStore defines the interface for managing mapping files
type MappingStore interface {
	// Load reads the mapping file from the specified path
	Load(path string) (*Mapping, error)

	// Save writes the mapping data to the specified path
	Save(path string, mapping *Mapping) error

	// Get returns the translation for a specific comment ID and language
	Get(commentID, lang string) (string, bool)

	// Set updates or adds a translation for a specific comment ID and language
	Set(commentID, lang, text string) error
}

// Mapping represents the structure of the mapping file
type Mapping struct {
	Version        string                       `json:"version"`
	SourceLanguage string                       `json:"sourceLanguage"`
	TargetLanguage string                       `json:"targetLanguage"`
	Comments       map[string]map[string]string `json:"comments"` // ID -> Lang -> Text
}
