package translator

import (
	"context"
	"fmt"
)

// MockTranslator is a dummy translator for testing
type MockTranslator struct{}

// NewMockTranslator creates a new MockTranslator
func NewMockTranslator() *MockTranslator {
	return &MockTranslator{}
}

// Translate returns a mock translation
func (t *MockTranslator) Translate(ctx context.Context, text, from, to string) (string, error) {
	return fmt.Sprintf("[MOCK %s->%s] %s", from, to, text), nil
}

// TranslateBatch returns mock translations
func (t *MockTranslator) TranslateBatch(ctx context.Context, texts []string, from, to string) ([]string, error) {
	results := make([]string, len(texts))
	for i, text := range texts {
		results[i] = fmt.Sprintf("[MOCK %s->%s] %s", from, to, text)
	}
	return results, nil
}
