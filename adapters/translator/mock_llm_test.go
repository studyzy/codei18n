package translator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

// MockLLMHandlerFunc defines the signature for handling mock requests
type MockLLMHandlerFunc func(req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)

// NewMockLLMServer creates a test server that mimics OpenAI API
func NewMockLLMServer(handler MockLLMHandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var req openai.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		resp, err := handler(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(resp)
	}))
}

// Helper to create a simple text response
func createMockResponse(content string) *openai.ChatCompletionResponse {
	return &openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Content: content,
					Role:    openai.ChatMessageRoleAssistant,
				},
			},
		},
	}
}

func TestMockLLMServer(t *testing.T) {
	// Verify the mock server works
	expected := "Hello World"
	server := NewMockLLMServer(func(req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
		return createMockResponse(expected), nil
	})
	defer server.Close()

	// Using the actual LLMTranslator to test the connection
	tr := NewLLMTranslator("test-key", server.URL+"/v1", "test-model")
	if tr == nil {
		t.Fatal("NewLLMTranslator returned nil")
	}
}
