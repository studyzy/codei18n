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
	translateBatchSize   int
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

	translateCmd.Flags().StringVar(&translateProvider, "provider", "", "覆盖配置文件中的提供商 (llm, openai, mock)")
	translateCmd.Flags().StringVar(&translateModel, "model", "", "指定模型 (如 gpt-3.5-turbo)")
	translateCmd.Flags().IntVar(&translateConcurrency, "concurrency", 5, "并发请求数")
	translateCmd.Flags().IntVar(&translateBatchSize, "batch-size", 0, "每批翻译的数量 (覆盖配置)")
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
	if translateBatchSize > 0 {
		cfg.BatchSize = translateBatchSize
	}

	// 2. Init Translator
	var trans core.Translator

	switch cfg.TranslationProvider {
	case "mock":
		trans = translator.NewMockTranslator()
	case "openai", "llm": // Support both openai and llm
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Fatal("未设置 OPENAI_API_KEY 环境变量")
		}

		// Debug config loading
		log.Info("Translation Config Loaded: %v", cfg.TranslationConfig)

		baseURL := os.Getenv("OPENAI_BASE_URL")

		// Case-insensitive check for baseUrl/baseURL/base_url
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

		log.Info("Using LLM: BaseURL=%s, Model=%s, BatchSize=%d", baseURL, model, cfg.BatchSize)
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

	log.Info("发现 %d 条待翻译注释，开始批量翻译 (BatchSize=%d, Concurrency=%d)...", len(tasks), cfg.BatchSize, translateConcurrency)

	// 5. Process with Batching and Concurrency
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " 正在翻译..."
	s.Writer = os.Stderr
	s.Start()

	var wg sync.WaitGroup
	sem := make(chan struct{}, translateConcurrency)
	successCount := 0
	failCount := 0
	var countMu sync.Mutex

	// Split tasks into batches
	batches := make([][]task, 0, (len(tasks)+cfg.BatchSize-1)/cfg.BatchSize)
	for i := 0; i < len(tasks); i += cfg.BatchSize {
		end := i + cfg.BatchSize
		if end > len(tasks) {
			end = len(tasks)
		}
		batches = append(batches, tasks[i:end])
	}

	for _, batch := range batches {
		wg.Add(1)
		sem <- struct{}{} // Acquire token

		go func(currentBatch []task) {
			defer wg.Done()
			defer func() { <-sem }() // Release token

			// Prepare batch input
			texts := make([]string, len(currentBatch))
			for i, t := range currentBatch {
				texts[i] = t.text
			}
			// Assume all tasks in a batch have same from/to (should be true if we sort/filter, but here tasks might be mixed directions?)
			// Wait, tasks loop iterates map, order undefined. It might mix En->Zh and Zh->En.
			// Batch translation requires same direction.
			// We need to group by direction first. But usually it's mostly one direction.
			// Let's safe guard: process batch only if direction matches.
			// Or better: Group batches by direction.
			// Current simple implementation:
			// If direction changes in batch, we can't batch efficiently.
			// Let's refine batching strategy.

			// Group by direction
			// Actually, just calling TranslateBatch with mixed tasks is wrong because TranslateBatch takes `from, to` args.
			// So we MUST ensure batch has same direction.
			// Given loop order is random, we should probably separate tasks by direction first.
			// But for now let's assume we can just check.

			// Refined logic: TranslateBatch takes single from/to.
			// So we must verify batch consistency.

			// Since we can't easily change the batching structure inside this goroutine loop without complicating it,
			// let's do grouping before creating batches.

			// Just use the first task's direction. If mixed, we might fail or need to split.
			// Actually, let's fix the task collection to be grouped.
			// But wait, I already split into batches.
			// The correct fix is to sort tasks or group tasks by direction before batching.
			// Let's just group by direction in the loop below.
			// BUT: simpler is just to ensure TranslateBatch is called with correct params.
			// If a batch has mixed directions, we can't use TranslateBatch easily.

			// Let's assume for typical usage (filling missing), it's usually En->Zh.
			// But mixed is possible.
			// I will add a check: if batch has mixed directions, fallback to sequential logic or split it?
			// Or better: Sort tasks by direction before batching.

			// Let's perform translation
			from := currentBatch[0].fromLang
			to := currentBatch[0].toLang
			consistent := true
			for _, t := range currentBatch {
				if t.fromLang != from || t.toLang != to {
					consistent = false
					break
				}
			}

			var results []string
			var err error

			if consistent {
				results, err = trans.TranslateBatch(context.Background(), texts, from, to)
			} else {
				// Mixed batch, fallback to sequential loop manually here
				// This shouldn't happen often if we sort, but safety net.
				results = make([]string, len(currentBatch))
				for i, t := range currentBatch {
					res, e := trans.Translate(context.Background(), t.text, t.fromLang, t.toLang)
					if e != nil {
						err = e // Capture last error
						break
					}
					results[i] = res
				}
			}

			countMu.Lock()
			defer countMu.Unlock()

			if err != nil {
				failCount += len(currentBatch)
			} else {
				// Save results
				for i, res := range results {
					t := currentBatch[i]
					store.Set(t.id, t.toLang, res)
					successCount++
				}
			}

			s.Suffix = fmt.Sprintf(" 正在翻译... (%d/%d 成功, %d 失败)", successCount, len(tasks), failCount)
		}(batch)
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
