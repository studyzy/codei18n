package translator

import (
	"context"
	"os"
	"testing"
	"time"
)

// 单元测试：验证 NewOllamaTranslator 默认值行为（不依赖本地服务）。
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

// 集成测试（可选）：连接本地 Ollama 服务并尝试真实翻译。
//
// 启用方式：
//
//	CODEI18N_OLLAMA_TEST=1 go test ./adapters/translator -run TestOllamaTranslator_Translate_Integration -v
//
// 可选环境变量：
//
//	CODEI18N_OLLAMA_ENDPOINT  默认为 http://localhost:11434
//	CODEI18N_OLLAMA_MODEL     默认为 qwen3:4b
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
