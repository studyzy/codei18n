package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OllamaTranslator 使用本地 Ollama 服务实现翻译。
// 通过 /api/chat 接口与本地模型交互。
type OllamaTranslator struct {
	endpoint   string
	model      string
	httpClient *http.Client
}

// NewOllamaTranslator 创建一个新的 OllamaTranslator。
// endpoint 例如 http://localhost:11434
// model 例如 "llama3"、"qwen2.5:7b" 等。
func NewOllamaTranslator(endpoint, model string) *OllamaTranslator {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3"
	}

	return &OllamaTranslator{
		endpoint: endpoint,
		model:    model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Translate 实现单条文本翻译。
func (t *OllamaTranslator) Translate(ctx context.Context, text, from, to string) (string, error) {
	prompt := fmt.Sprintf(
		"You are a professional code comment translator. Translate the following code comment from %s to %s.\n"+
			"Rules:\n"+
			"1. Keep technical terms, variable names, and code snippets unchanged.\n"+
			"2. Maintain the tone and style of the original comment.\n"+
			"3. Output ONLY the translated text, no explanations or quotes.\n"+
			"4. If the text is already in the target language, return it as is.\n"+
			"5. Preserve all line breaks and formatting.\n\n"+
			"Original: %s",
		from, to, text,
	)

	reqBody := struct {
		Model    string          `json:"model"`
		Messages []ollamaMessage `json:"messages"`
		Stream   bool            `json:"stream"`
		// think 控制思维链模式，对于 DeepSeek/Qwen 等思考模型，显式关闭可避免额外消耗
		Think   bool              `json:"think"`
		Options map[string]string `json:"options,omitempty"`
	}{
		Model: t.model,
		Messages: []ollamaMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
		Think:  false,
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint+"/api/chat", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama 请求失败: %s", resp.Status)
	}

	var respBody struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}

	if respBody.Message.Content == "" {
		return "", fmt.Errorf("ollama 返回内容为空")
	}

	return strings.TrimSpace(respBody.Message.Content), nil
}

// TranslateBatch 在 Ollama 下采用顺序调用实现，避免过早引入复杂批量协议。
func (t *OllamaTranslator) TranslateBatch(ctx context.Context, texts []string, from, to string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}
	if len(texts) == 1 {
		res, err := t.Translate(ctx, texts[0], from, to)
		if err != nil {
			return nil, err
		}
		return []string{res}, nil
	}

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

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
