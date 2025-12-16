package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/studyzy/codei18n/adapters/translator"
	"github.com/studyzy/codei18n/core"
	"github.com/studyzy/codei18n/core/config"
	"github.com/studyzy/codei18n/core/mapping"
	"github.com/studyzy/codei18n/internal/log"
)

var (
	translateProvider    string
	translateModel       string
	translateConcurrency int
)

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "自动翻译缺失的注释",
	Long:  `调用配置的翻译引擎（Google/OpenAI/DeepSeek）自动翻译映射文件中缺失的条目。`,
	Run: func(cmd *cobra.Command, args []string) {
		runTranslate()
	},
}

func init() {
	rootCmd.AddCommand(translateCmd)

	translateCmd.Flags().StringVar(&translateProvider, "provider", "", "覆盖配置文件中的提供商 (google, openai, mock)")
	translateCmd.Flags().StringVar(&translateModel, "model", "", "指定模型 (如 gpt-3.5-turbo)")
	translateCmd.Flags().IntVar(&translateConcurrency, "concurrency", 5, "并发请求数")
}

func runTranslate() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Warn("无法加载配置: %v", err)
		cfg = config.DefaultConfig()
	}

	// Override flags
	if translateProvider != "" {
		cfg.TranslationProvider = translateProvider
	}
	// Model overrides? We need to verify if provider supports model config
	// Currently stored in map

	// 2. Init Translator
	var trans core.Translator

	switch cfg.TranslationProvider {
	case "mock":
		trans = translator.NewMockTranslator()
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Fatal("未设置 OPENAI_API_KEY 环境变量")
		}

		// Debug config loading
		log.Info("Translation Config Loaded: %v", cfg.TranslationConfig)

		baseURL := os.Getenv("OPENAI_BASE_URL")

		// Case-insensitive check for baseUrl/baseURL/base_url
		// Priority: config file > environment variable
		if val, ok := cfg.TranslationConfig["baseUrl"]; ok && val != "" {
			baseURL = val
		} else if val, ok := cfg.TranslationConfig["BaseUrl"]; ok && val != "" {
			baseURL = val
		} else if val, ok := cfg.TranslationConfig["baseURL"]; ok && val != "" {
			baseURL = val
		} else if val, ok := cfg.TranslationConfig["base_url"]; ok && val != "" {
			baseURL = val
		} else if val, ok := cfg.TranslationConfig["baseurl"]; ok && val != "" {
			// Supports fully lowercase baseurl
			baseURL = val
		}

		model := "gpt-3.5-turbo"
		if translateModel != "" {
			model = translateModel
		} else if m, ok := cfg.TranslationConfig["model"]; ok {
			model = m
		}

		// Auto-detect DeepSeek URL if not set
		if baseURL == "" && (model == "deepseek-chat" || model == "deepseek-coder") {
			baseURL = "https://api.deepseek.com"
			log.Info("自动检测到 DeepSeek 模型，设置 BaseURL 为 %s", baseURL)
		}

		log.Info("Using LLM: BaseURL=%s, Model=%s", baseURL, model)
		trans = translator.NewLLMTranslator(apiKey, baseURL, model)
	default:
		log.Fatal("不支持的翻译提供商: %s", cfg.TranslationProvider)
	}

	// 3. Load Mapping
	storePath := filepath.Join(".codei18n", "mappings.json")
	store := mapping.NewStore(storePath)
	if err := store.Load(); err != nil {
		log.Fatal("加载映射文件失败: %v", err)
	}

	m := store.GetMapping()

	// 4. Identify missing translations
	// Strategy: Bidirectional Translation
	// - If en exists but zh-CN is missing: en -> zh-CN
	// - If zh-CN exists but en is missing: zh-CN -> en
	type task struct {
		id       string
		text     string
		fromLang string
		toLang   string
	}
	var tasks []task

	for id, translations := range m.Comments {
		// Case 1: EN exists, ZH missing -> Translate EN to ZH
		if enText, hasEn := translations[cfg.SourceLanguage]; hasEn && enText != "" {
			if zhText, hasZh := translations[cfg.LocalLanguage]; !hasZh || zhText == "" {
				tasks = append(tasks, task{
					id:       id,
					text:     enText,
					fromLang: cfg.SourceLanguage,
					toLang:   cfg.LocalLanguage,
				})
			}
		}

		// Case 2: ZH exists, EN missing -> Translate ZH to EN (reverse translation)
		if zhText, hasZh := translations[cfg.LocalLanguage]; hasZh && zhText != "" {
			if enText, hasEn := translations[cfg.SourceLanguage]; !hasEn || enText == "" {
				tasks = append(tasks, task{
					id:       id,
					text:     zhText,
					fromLang: cfg.LocalLanguage,
					toLang:   cfg.SourceLanguage,
				})
			}
		}
	}

	if len(tasks) == 0 {
		log.Success("所有注释已翻译，无需操作")
		return
	}

	log.Info("发现 %d 条待翻译注释，开始翻译...", len(tasks))

	// 5. Process with concurrency
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " 正在翻译..."
	s.Writer = os.Stderr // Spinner to stderr!
	s.Start()

	var wg sync.WaitGroup
	sem := make(chan struct{}, translateConcurrency)
	successCount := 0
	failCount := 0

	// Mutex for counting
	var countMu sync.Mutex

	for _, t := range tasks {
		wg.Add(1)
		sem <- struct{}{} // Acquire token

		go func(t task) {
			defer wg.Done()
			defer func() { <-sem }() // Release token

			res, err := trans.Translate(context.Background(), t.text, t.fromLang, t.toLang)

			countMu.Lock()
			defer countMu.Unlock()

			if err != nil {
				// log.Warn("翻译失败 [%s]: %v", t.id, err) // Avoid spamming log during spinner
				failCount++
			} else {
				store.Set(t.id, t.toLang, res)
				successCount++
			}

			// Update spinner
			s.Suffix = fmt.Sprintf(" 正在翻译... (%d/%d 成功, %d 失败)", successCount, len(tasks), failCount)
		}(t)
	}

	wg.Wait()
	s.Stop()

	// 6. Save
	if err := store.Save(); err != nil {
		log.Fatal("保存映射文件失败: %v", err)
	}

	if failCount > 0 {
		log.Warn("翻译完成，但有 %d 条失败", failCount)
	} else {
		log.Success("翻译完成！共处理 %d 条注释", successCount)
	}
}
