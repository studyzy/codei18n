package translator

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslateNewlinePreservation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}
	baseURL := os.Getenv("OPENAI_BASE_URL")
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	// Input with newlines
	input := "Line 1\nLine 2\nLine 3"

	translator := NewLLMTranslator(apiKey, baseURL, model)

	// 1. Test Single Translate
	t.Run("Single Translate", func(t *testing.T) {
		result, err := translator.Translate(context.Background(), input, "en", "zh-CN")
		require.NoError(t, err)

		// Check if newlines are preserved
		assert.Contains(t, result, "\n", "Result should contain newlines")
		lines := strings.Split(result, "\n")
		// We expect roughly the same number of lines
		assert.GreaterOrEqual(t, len(lines), 3, "Result should have at least 3 lines")
		t.Logf("Single Translate Result:\n%s", result)
	})

	// 2. Test Batch Translate
	t.Run("Batch Translate", func(t *testing.T) {
		results, err := translator.TranslateBatch(context.Background(), []string{input}, "en", "zh-CN")
		require.NoError(t, err)
		require.Len(t, results, 1)

		// Check if newlines are preserved
		assert.Contains(t, results[0], "\n", "Result should contain newlines")
		lines := strings.Split(results[0], "\n")
		assert.GreaterOrEqual(t, len(lines), 3, "Result should have at least 3 lines")
		t.Logf("Batch Translate Result:\n%s", results[0])
	})
}
