package translator

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// LLMTranslator implements Translator using OpenAI compatible API
type LLMTranslator struct {
	client *openai.Client
	model  string
}

// NewLLMTranslator creates a new translator.
// apiKey: API Key
// baseURL: Optional base URL (for DeepSeek/others). If empty, uses OpenAI default.
// model: Model name (e.g. gpt-3.5-turbo, deepseek-chat)
func NewLLMTranslator(apiKey, baseURL, model string) *LLMTranslator {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)
	return &LLMTranslator{
		client: client,
		model:  model,
	}
}

// Translate translates a single text
func (t *LLMTranslator) Translate(ctx context.Context, text, from, to string) (string, error) {
	// Simple prompt construction
	prompt := fmt.Sprintf(
		"You are a professional code comment translator. Translate the following code comment from %s to %s.\n"+
			"Rules:\n"+
			"1. Keep technical terms, variable names, and code snippets unchanged.\n"+
			"2. Maintain the tone and style of the original comment.\n"+
			"3. Output ONLY the translated text, no explanations or quotes.\n"+
			"4. If the text is already in the target language, return it as is.\n\n"+
			"Original: %s",
		from, to, text,
	)

	resp, err := t.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: t.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from LLM")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// TranslateBatch translates a batch of texts
// Implementation strategy:
//  1. Send all texts in one prompt (if length permits) or multiple concurrent requests.
//  2. For simplicity and robustness against context limits, we'll use concurrent single requests for now,
//     or a simple JSON list approach in prompt.
//     JSON list approach saves tokens and network RTT.
func (t *LLMTranslator) TranslateBatch(ctx context.Context, texts []string, from, to string) ([]string, error) {
	// Fallback to sequential for now to ensure reliability,
	// unless we implement a robust JSON array parser for the response.
	// Since the Interface allows it, the caller can implement concurrency.
	// But let's try a simple loop here. Ideally, `translate` command handles concurrency.

	results := make([]string, len(texts))
	for i, text := range texts {
		res, err := t.Translate(ctx, text, from, to)
		if err != nil {
			return nil, err
		}
		results[i] = res
	}
	return results, nil
}
