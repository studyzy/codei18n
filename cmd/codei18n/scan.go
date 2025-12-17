package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/studyzy/codei18n/adapters"
	"github.com/studyzy/codei18n/core/config"
	"github.com/studyzy/codei18n/core/domain"
	"github.com/studyzy/codei18n/core/mapping"
	"github.com/studyzy/codei18n/core/utils"
	"github.com/studyzy/codei18n/internal/log"
)

var (
	scanFile             string
	scanDir              string
	scanFormat           string
	scanOutput           string
	scanLang             string
	scanStdin            bool
	scanWithTranslations bool
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "扫描源码并提取注释",
	Long: `扫描指定的源码文件或目录，提取注释信息并生成唯一 ID。
支持多种输出格式（JSON, Table），并支持从 Stdin 读取内容以支持 IDE 集成。`,
	Run: func(cmd *cobra.Command, args []string) {
		runScan()
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringVarP(&scanFile, "file", "f", "", "指定扫描的单个文件路径")
	scanCmd.Flags().StringVarP(&scanDir, "dir", "d", ".", "指定扫描的目录路径")
	scanCmd.Flags().StringVar(&scanFormat, "format", "table", "输出格式 (json, table)")
	scanCmd.Flags().StringVarP(&scanOutput, "output", "o", "", "将输出写入指定文件 (默认 stdout)")
	scanCmd.Flags().StringVar(&scanLang, "lang", "", "指定目标语言 (覆盖配置)")
	scanCmd.Flags().BoolVar(&scanStdin, "stdin", false, "从 stdin 读取文件内容 (必须同时指定 --file)")
	scanCmd.Flags().BoolVar(&scanWithTranslations, "with-translations", false, "在 JSON 输出中包含翻译文本")
}

func runScan() {
	// Validation
	if scanStdin && scanFile == "" {
		log.Fatal("--stdin 模式下必须指定 --file 以确定语言和 ID 上下文")
	}

	var comments []*domain.Comment
	var err error

	// Determine strategy
	if scanStdin {
		comments, err = scanFromStdin(scanFile)
	} else if scanFile != "" {
		comments, err = scanSingleFile(scanFile)
	} else {
		comments, err = scanDirectory(scanDir)
	}

	if err != nil {
		log.Fatal("扫描失败: %v", err)
	}

	// Calculate IDs for all comments
	for _, c := range comments {
		if c.ID == "" {
			c.ID = utils.GenerateCommentID(c)
		}
	}

	// Load mapping if requested
	if scanWithTranslations {
		// Load config
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Warn("无法加载配置，使用默认值: %v", err)
			cfg = config.DefaultConfig()
		}

		// Load store
		storePath := filepath.Join(".codei18n", "mappings.json")
		store := mapping.NewStore(storePath)
		// We ignore error here, if file doesn't exist, we just don't show translations
		// or better: log.Warn
		if err := store.Load(); err != nil {
			log.Warn("加载映射文件失败: %v", err)
		} else {
			// Populate LocalizedText
			targetLang := cfg.LocalLanguage // Default target
			if scanLang != "" {
				targetLang = scanLang
			}
			for _, c := range comments {
				if text, ok := store.Get(c.ID, targetLang); ok {
					c.LocalizedText = text
				}
			}
		}
	}

	// Output
	if err := outputResults(comments); err != nil {
		log.Fatal("输出结果失败: %v", err)
	}
}

func scanFromStdin(filename string) ([]*domain.Comment, error) {
	// Read stdin
	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("读取 stdin 失败: %w", err)
	}

	adapter, err := adapters.GetAdapter(filename)
	if err != nil {
		return nil, err
	}
	return adapter.Parse(filename, src)
}

func scanSingleFile(filename string) ([]*domain.Comment, error) {
	adapter, err := adapters.GetAdapter(filename)
	if err != nil {
		return nil, err
	}

	// Read file content manually to ensure consistency with directory scan
	// and to debug potential read issues
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	return adapter.Parse(filename, content)
}

func scanDirectory(dir string) ([]*domain.Comment, error) {
	var comments []*domain.Comment
	var walkErr error

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip .git, vendor, .codei18n
			if info.Name() == ".git" || info.Name() == "vendor" || info.Name() == ".codei18n" {
				return filepath.SkipDir
			}
			return nil
		}

		// Try to get adapter for file
		adapter, err := adapters.GetAdapter(path)
		if err == nil {
			// Supported file
			// Calculate relative path for ID stability
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				relPath = path
			}

			// Read file content manually to ensure we access the correct file
			content, err := os.ReadFile(path)
			if err != nil {
				log.Warn("读取文件 %s 失败: %v", path, err)
				return nil
			}

			fileComments, err := adapter.Parse(relPath, content)
			if err != nil {
				log.Warn("解析文件 %s 失败: %v", path, err)
				return nil // Continue scanning other files
			}
			comments = append(comments, fileComments...)
		}
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}
	return comments, err
}

func outputResults(comments []*domain.Comment) error {
	// Prepare output data
	type Output struct {
		File     string            `json:"file,omitempty"` // populated if single file? Or just list?
		Comments []*domain.Comment `json:"comments"`
	}

	var data interface{}
	if scanFile != "" || scanStdin {
		// Single file context
		data = Output{
			File:     scanFile,
			Comments: comments,
		}
	} else {
		// Directory context - just return the list
		data = comments
	}

	if scanFormat == "json" {
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}

		if scanOutput != "" {
			return os.WriteFile(scanOutput, jsonData, 0644)
		}

		log.PrintJSON(jsonData)
		return nil
	}

	// Table format (for humans)
	if scanOutput != "" {
		// Write simple text to file
		f, err := os.Create(scanOutput)
		if err != nil {
			return err
		}
		defer f.Close()
		for _, c := range comments {
			fmt.Fprintf(f, "[%s] %s: %s\n", c.ID[:8], c.Symbol, c.SourceText)
		}
		return nil
	}

	// Print to stdout
	fmt.Printf("扫描完成，共找到 %d 条注释:\n", len(comments))
	for _, c := range comments {
		fmt.Printf("- [%s] %s (%d:%d): %s\n", c.ID[:8], c.Symbol, c.Range.StartLine, c.Range.StartCol, c.SourceText)
	}
	return nil
}
