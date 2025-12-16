package main

import (
	"github.com/spf13/cobra"
	"github.com/studyzy/codei18n/core/config"
	"github.com/studyzy/codei18n/internal/log"
)

var (
	initSourceLang string
	initTargetLang string
	initProvider   string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化项目配置",
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initSourceLang, "source-lang", "en", "源码语言")
	initCmd.Flags().StringVar(&initTargetLang, "target-lang", "zh-CN", "本地目标语言")
	initCmd.Flags().StringVar(&initProvider, "provider", "google", "翻译提供商")
}

func runInit() {
	cfg := config.DefaultConfig()
	cfg.SourceLanguage = initSourceLang
	cfg.LocalLanguage = initTargetLang
	cfg.TranslationProvider = initProvider

	if err := config.SaveConfig(cfg); err != nil {
		log.Fatal("保存配置失败: %v", err)
	}

	log.Success("项目初始化成功")
}
