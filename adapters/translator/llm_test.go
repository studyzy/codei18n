package translator

import (
	"context"
	"fmt"
	"os"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBatchTranslationMock tests batch translation using a mock server
func TestBatchTranslationMock(t *testing.T) {
	// 1. Success Scenario
	successHandler := func(req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
		// Verify Prompt contains JSON
		content := req.Messages[0].Content
		if !contains(content, `["Hello","World"]`) {
			return nil, fmt.Errorf("unexpected prompt: %s", content)
		}

		// Return valid JSON array
		return createMockResponse(`["你好", "世界"]`), nil
	}

	server1 := NewMockLLMServer(successHandler)
	defer server1.Close()

	trans1 := NewLLMTranslator("key", server1.URL, "model")
	results, err := trans1.TranslateBatch(context.Background(), []string{"Hello", "World"}, "en", "zh-CN")

	require.NoError(t, err)
	assert.Equal(t, []string{"你好", "世界"}, results)

	// 2. Fallback Scenario (JSON Parse Error)
	fallbackHandler := func(req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
		content := req.Messages[0].Content
		// If batch prompt
		if contains(content, "JSON array") {
			return createMockResponse("Not a JSON response"), nil
		}
		// If sequential fallback (single item)
		if contains(content, "Original: Hello") {
			return createMockResponse("你好"), nil
		}
		if contains(content, "Original: World") {
			return createMockResponse("世界"), nil
		}
		return nil, fmt.Errorf("unknown prompt")
	}

	server2 := NewMockLLMServer(fallbackHandler)
	defer server2.Close()

	trans2 := NewLLMTranslator("key", server2.URL, "model")
	results2, err := trans2.TranslateBatch(context.Background(), []string{"Hello", "World"}, "en", "zh-CN")

	require.NoError(t, err)
	assert.Equal(t, []string{"你好", "世界"}, results2)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}

// TestDeepSeekIntegration is a manual test to verify API connectivity.
// It runs only if OPENAI_API_KEY is set and -short flag is NOT used.
// Run with: go test -run TestDeepSeekIntegration (without -short)
func TestDeepSeekIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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

	trans := NewLLMTranslator(apiKey, baseURL, model)

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

func TestEdgeCases(t *testing.T) {
	// Test empty array
	server := NewMockLLMServer(func(req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
		return createMockResponse(`[]`), nil
	})
	defer server.Close()

	trans := NewLLMTranslator("key", server.URL, "model")
	res, err := trans.TranslateBatch(context.Background(), []string{}, "en", "zh")
	assert.NoError(t, err)
	assert.Empty(t, res)

	// Test special characters
	specials := []string{`Quote "`, `Newline \n`}
	server2 := NewMockLLMServer(func(req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
		// Naive mock returning input as if translated
		return createMockResponse(`["Quote \"", "Newline \\n"]`), nil
	})
	defer server2.Close()

	trans2 := NewLLMTranslator("key", server2.URL, "model")
	res2, err := trans2.TranslateBatch(context.Background(), specials, "en", "zh")
	assert.NoError(t, err)
	assert.Equal(t, specials, res2)
}
