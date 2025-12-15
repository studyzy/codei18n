package translator_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/studyzy/codei18n/adapters/translator"
)

// TestDeepSeekIntegration is a manual test to verify API connectivity.
// It runs only if OPENAI_API_KEY is set.
func TestDeepSeekIntegration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	// Read optional config from env or default to DeepSeek
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com" // Default DeepSeek base
	}
	
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}

	fmt.Printf("Testing with:\nBaseURL: %s\nModel: %s\n", baseURL, model)

	trans := translator.NewLLMTranslator(apiKey, baseURL, model)

	text := "Calculate the sum of two numbers"
	from := "en"
	to := "zh-CN"

	fmt.Printf("Translating: '%s' (%s -> %s)...\n", text, from, to)

	result, err := trans.Translate(context.Background(), text, from, to)
	
	if err != nil {
		t.Fatalf("Translation failed: %v", err)
	}

	fmt.Printf("Success! Result: %s\n", result)
	
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	// Basic check to see if it looks like Chinese (or at least changed)
	assert.NotEqual(t, text, result)
}
