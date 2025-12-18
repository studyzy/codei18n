package translator

import (
	"context"
	"os"
	"testing"
	"time"
)

// Unit test: Verify the default value behavior of NewOllamaTranslator (without relying on local service).
func TestNewOllamaTranslator_Defaults(t *testing.T) {
	tr := NewOllamaTranslator("", "")
	if tr.endpoint == "" {
		t.Fatalf("expected non-empty default endpoint, got empty")
	}
	if tr.model == "" {
		t.Fatalf("expected non-empty default model, got empty")
	}
	if tr.httpClient == nil {
		t.Fatalf("expected non-nil httpClient")
	}
}

// Integration test (optional): Connect to local Ollama service and attempt real translation.
//
// How to enable:
//
//	CODEI18N_OLLAMA_TEST=1 go test ./adapters/translator -run TestOllamaTranslator_Translate_Integration -v
//
// Optional environment variables:
//
//	CODEI18N_OLLAMA_ENDPOINT  Defaults to http://localhost:11434
//	CODEI18N_OLLAMA_MODEL     Defaults to qwen3:4b
func TestOllamaTranslator_Translate_Integration(t *testing.T) {
	if os.Getenv("CODEI18N_OLLAMA_TEST") == "" {
		t.Skip("跳过 Ollama 集成测试：未设置 CODEI18N_OLLAMA_TEST")
	}

	endpoint := os.Getenv("CODEI18N_OLLAMA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	model := os.Getenv("CODEI18N_OLLAMA_MODEL")
	if model == "" {
		model = "qwen3:4b"
	}

	tr := NewOllamaTranslator(endpoint, model)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	text := "Calculate the sum of two numbers"
	res, err := tr.Translate(ctx, text, "en", "zh-CN")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	if res == "" {
		t.Fatalf("expected non-empty translation result")
	}
}
