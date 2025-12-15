package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/studyzy/codei18n/core/config"
	"github.com/studyzy/codei18n/core/mapping"
	"github.com/studyzy/codei18n/core/utils"
	"github.com/studyzy/codei18n/internal/log"
)

var (
	mapScanDir string
	mapDryRun  bool
	mapLang    string
)

// mapCmd represents the map command
var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "管理多语言映射",
	Long:  `创建、更新或查询注释的多语言映射文件。`,
}

// mapUpdateCmd represents the map update command
var mapUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新映射文件",
	Long:  `扫描项目中的注释，并将新发现的注释添加到映射文件中。`,
	Run: func(cmd *cobra.Command, args []string) {
		runMapUpdate()
	},
}

// mapGetCmd represents the map get command
var mapGetCmd = &cobra.Command{
	Use:   "get [commentID]",
	Short: "查询注释翻译",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMapGet(args[0])
	},
}

func init() {
	rootCmd.AddCommand(mapCmd)
	mapCmd.AddCommand(mapUpdateCmd)
	mapCmd.AddCommand(mapGetCmd)

	mapUpdateCmd.Flags().StringVar(&mapScanDir, "scan-dir", ".", "扫描目录以更新映射")
	mapUpdateCmd.Flags().BoolVar(&mapDryRun, "dry-run", false, "仅显示变更，不写入文件")

	mapGetCmd.Flags().StringVar(&mapLang, "lang", "", "目标语言代码 (默认使用配置中的 LocalLanguage)")
}

func runMapUpdate() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Warn("无法加载配置，使用默认值: %v", err)
		cfg = config.DefaultConfig()
	}

	// 1. Scan for comments
	log.Info("正在扫描目录: %s", mapScanDir)
	comments, err := scanDirectory(mapScanDir)
	if err != nil {
		log.Fatal("扫描失败: %v", err)
	}

	// 2. Generate IDs
	for _, c := range comments {
		if c.ID == "" {
			c.ID = utils.GenerateCommentID(c)
		}
	}

	// 3. Load Store
	storePath := filepath.Join(".codei18n", "mappings.json")
	store := mapping.NewStore(storePath)
	if err := store.Load(); err != nil {
		log.Fatal("加载映射文件失败: %v", err)
	}

	// 4. Update
	m := store.GetMapping()
	m.SourceLanguage = cfg.SourceLanguage
	m.TargetLanguage = cfg.LocalLanguage

	addedCount := 0
	for _, c := range comments {
		// Check if ID exists
		if _, exists := m.Comments[c.ID]; !exists {
			store.Set(c.ID, cfg.SourceLanguage, c.SourceText)
			// Initialize target lang as empty string (or maybe not?)
			// Ideally we just set the source.
			// The Set method does: m.Comments[id][lang] = text
			// We want to ensure the entry exists.
			addedCount++
		}
	}

	log.Success("发现 %d 条注释，新增 %d 条映射", len(comments), addedCount)

	if mapDryRun {
		log.Info("Dry run 模式，不保存文件")
		return
	}

	// 5. Save
	if err := store.Save(); err != nil {
		log.Fatal("保存映射文件失败: %v", err)
	}
	log.Success("映射文件已更新: %s", storePath)
}

func runMapGet(commentID string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	targetLang := mapLang
	if targetLang == "" {
		targetLang = cfg.LocalLanguage
	}

	storePath := filepath.Join(".codei18n", "mappings.json")
	store := mapping.NewStore(storePath)
	if err := store.Load(); err != nil {
		log.Fatal("加载映射文件失败: %v", err)
	}

	if text, ok := store.Get(commentID, targetLang); ok {
		fmt.Println(text)
	} else {
		log.Error("未找到 ID 为 %s 的 %s 翻译", commentID, targetLang)
		os.Exit(1)
	}
}
