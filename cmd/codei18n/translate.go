package main

import (
	"github.com/spf13/cobra"
	"github.com/studyzy/codei18n/core/config"
	"github.com/studyzy/codei18n/core/workflow"
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
	Long:  `调用配置的翻译引擎（LLM(OpenAI/DeepSeek) 或本地 Ollama）自动翻译映射文件中缺失的条目。`,
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

	opts := workflow.TranslateOptions{
		Concurrency: translateConcurrency,
		Provider:    translateProvider,
		Model:       translateModel,
		BatchSize:   translateBatchSize,
	}

	result, err := workflow.Translate(cfg, opts)
	if err != nil {
		log.Fatal("Translation failed: %v", err)
	}

	if result.TotalTasks == 0 {
		log.Success("所有注释已翻译，无需操作")
		return
	}

	if result.FailCount > 0 {
		log.Warn("翻译完成，但有 %d 条失败", result.FailCount)
	} else {
		log.Success("翻译完成！共处理 %d 条注释", result.SuccessCount)
	}
}
