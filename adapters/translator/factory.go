package translator

import (
	"fmt"
	"os"
	"strings"

	"github.com/studyzy/codei18n/core"
	"github.com/studyzy/codei18n/core/config"
	"github.com/studyzy/codei18n/internal/log"
)

// NewFromConfig 根据配置创建合适的翻译引擎实现。
//
// 约定：
//   - provider 为 "openai" 或 "llm" 时，使用基于 OpenAI 兼容协议的 LLM 翻译
//   - provider 为 "ollama" 时，使用本地 Ollama 服务
//   - provider 为 "mock" 时，使用测试用的 MockTranslator
//   - provider 为 "google" 或 "deepl" 时，视为已废弃并返回错误
//   - 若 provider 为空，则默认视为 "openai"
func NewFromConfig(cfg *config.Config) (core.Translator, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.TranslationProvider))
	if provider == "" {
		provider = "openai"
	}

	// 兼容文档中可能出现的 llm-api 命名
	if provider == "llm-api" {
		provider = "openai"
	}

	switch provider {
	case "google", "deepl":
		return nil, fmt.Errorf("翻译提供商 %q 已不再支持，请修改配置为 \"openai\" 或 \"ollama\"", provider)
	case "mock":
		return NewMockTranslator(), nil
	case "openai", "llm":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("未设置 OPENAI_API_KEY 环境变量")
		}

		baseURL := os.Getenv("OPENAI_BASE_URL")
		if cfg.TranslationConfig != nil {
			// 不同大小写和命名的兼容
			if val, ok := cfg.TranslationConfig["baseUrl"]; ok && val != "" {
				baseURL = val
			} else if val, ok := cfg.TranslationConfig["BaseUrl"]; ok && val != "" {
				baseURL = val
			} else if val, ok := cfg.TranslationConfig["baseURL"]; ok && val != "" {
				baseURL = val
			} else if val, ok := cfg.TranslationConfig["base_url"]; ok && val != "" {
				baseURL = val
			} else if val, ok := cfg.TranslationConfig["baseurl"]; ok && val != "" {
				baseURL = val
			}
		}

		model := "gpt-3.5-turbo"
		if cfg.TranslationConfig != nil {
			if m, ok := cfg.TranslationConfig["model"]; ok && m != "" {
				model = m
			}
		}

		// DeepSeek 模型的默认 BaseURL 自动检测
		if baseURL == "" && (model == "deepseek-chat" || model == "deepseek-coder") {
			baseURL = "https://api.deepseek.com"
			log.Info("自动检测到 DeepSeek 模型，设置 BaseURL 为 %s", baseURL)
		}

		log.Info("Using LLM: BaseURL=%s, Model=%s, BatchSize=%d", baseURL, model, cfg.BatchSize)
		return NewLLMTranslator(apiKey, baseURL, model), nil
	case "ollama":
		endpoint := "http://localhost:11434"
		model := "llama3"
		if cfg.TranslationConfig != nil {
			if v, ok := cfg.TranslationConfig["endpoint"]; ok && v != "" {
				endpoint = v
			}
			if m, ok := cfg.TranslationConfig["model"]; ok && m != "" {
				model = m
			}
		}
		log.Info("Using Ollama: Endpoint=%s, Model=%s, BatchSize=%d", endpoint, model, cfg.BatchSize)
		return NewOllamaTranslator(endpoint, model), nil
	default:
		return nil, fmt.Errorf("不支持的翻译提供商: %s", provider)
	}
}
