package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/studyzy/codei18n/adapters/golang"
	"github.com/studyzy/codei18n/core/config"
	"github.com/studyzy/codei18n/core/domain"
	"github.com/studyzy/codei18n/core/mapping"
	"github.com/studyzy/codei18n/core/utils"
	"github.com/studyzy/codei18n/internal/log"
)

var (
	convertFile   string
	convertDir    string
	convertDryRun bool
	convertTo     string
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "转换源码中的注释语言",
	Long:  `在源码中原地修改注释，将其替换为目标语言的翻译文本（或还原为源语言）。`,
	Run: func(cmd *cobra.Command, args []string) {
		runConvert()
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVarP(&convertFile, "file", "f", "", "指定文件")
	convertCmd.Flags().StringVarP(&convertDir, "dir", "d", ".", "指定目录")
	convertCmd.Flags().StringVar(&convertTo, "to", "", "目标语言 (en 或 zh-CN)")
	convertCmd.Flags().BoolVar(&convertDryRun, "dry-run", false, "仅显示将要修改的内容")
}

func runConvert() {
	if convertTo == "" {
		log.Fatal("必须指定目标语言: --to <lang>")
	}

	// Load Config & Mapping
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Warn("加载配置失败: %v", err)
		cfg = config.DefaultConfig()
	}

	storePath := filepath.Join(".codei18n", "mappings.json")
	store := mapping.NewStore(storePath)
	if err := store.Load(); err != nil {
		log.Fatal("加载映射文件失败: %v", err)
	}
	log.Info("Loaded store with %d comments", len(store.GetMapping().Comments))

	// Identify files
	var files []string
	if convertFile != "" {
		files = append(files, convertFile)
	} else {
		err := filepath.Walk(convertDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if info.Name() == ".git" || info.Name() == "vendor" || info.Name() == ".codei18n" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".go") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			log.Fatal("扫描目录失败: %v", err)
		}
	}

	log.Info("准备处理 %d 个文件...", len(files))

	adapter := golang.NewAdapter()

	for _, file := range files {
		processFile(file, adapter, store, cfg)
	}

	if err := store.Save(); err != nil {
		log.Error("保存映射文件失败: %v", err)
	} else {
		log.Info("Mapping store updated.")
	}
}

func processFile(file string, adapter *golang.Adapter, store *mapping.Store, cfg *config.Config) {
	// Read file
	src, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error("读取文件 %s 失败: %v", file, err)
		return
	}

	// Parse to get comments
	comments, err := adapter.Parse(file, src)
	if err != nil {
		log.Error("解析文件 %s 失败: %v", file, err)
		return
	}

	log.Info("Convert: To='%s', Source='%s', Local='%s'", convertTo, cfg.SourceLanguage, cfg.LocalLanguage)

	// Filter comments that need update
	type replacement struct {
		startOffset int
		endOffset   int
		newText     string
	}
	var replacements []replacement

	lines := strings.Split(string(src), "\n")

	// Calculate line offsets map to translate Line:Col to ByteOffset
	lineOffsets := make([]int, len(lines))
	offset := 0
	for i, line := range lines {
		lineOffsets[i] = offset
		offset += len(line) + 1 // +1 for newline
	}

	for _, c := range comments {
		// Generate ID
		if c.ID == "" {
			c.ID = utils.GenerateCommentID(c)
		}

		log.Info("Checking ID: %s", c.ID)

		// Find target text
		var targetText string
		var found bool

		// Normalize current comment text for comparison
		normalizedCurrent := utils.NormalizeCommentText(c.SourceText)
		log.Info("Current Text: '%s' (normalized: '%s')", c.SourceText, normalizedCurrent)

		// If converting to SourceLanguage (e.g. en), we try to restore original
		if convertTo == cfg.SourceLanguage {
			// Restore Mode: ZH -> EN
			// Current text is likely in LocalLanguage (ZH)
			// Search in store by comparing normalized values
			for id, transMap := range store.GetMapping().Comments {
				zhText, hasZh := transMap[cfg.LocalLanguage]
				enText, hasEn := transMap[cfg.SourceLanguage]
				
				if hasZh && hasEn {
					normalizedZh := utils.NormalizeCommentText(zhText)
					if normalizedZh == normalizedCurrent {
						targetText = enText
						found = true
						log.Info("Found by reverse lookup: ID=%s, ZH='%s' -> EN='%s'", id, zhText, enText)

						// Key fix: When reverting to English, automatically generate IDs for new English comments and migrate existing translations to them.
						// This way, subsequent scan/map updates can directly recognize this English comment and will not treat it as an untranslated new comment.
						
						// 1. Construct the English version of the Comment object (simulating the converted state)
						tempC := *c
						tempC.SourceText = enText // Assuming mapping stores plain text or marked text, Normalize will process it
						
						// 2. Calculate new ID
						newID := utils.GenerateCommentID(&tempC)
						
						// 3. If the new ID differs from the old ID (it definitely will, because the text has changed), then migrate the data
						if newID != id {
							log.Info("Migrating mapping for restored English comment: %s -> %s", id, newID)
							// Ensure the new ID has bilingual data
							store.Set(newID, cfg.SourceLanguage, enText)
							store.Set(newID, cfg.LocalLanguage, zhText)
						}

						break
					}
				}
			}
			
			if !found {
				log.Warn("未找到注释的英文翻译: '%s'", normalizedCurrent)
			}
		} else {
			// Apply Mode (EN -> ZH)
			// c.SourceText is EN
			// ID generation works.
			if val, ok := store.Get(c.ID, convertTo); ok && val != "" {
				targetText = val
				found = true
				log.Info("Found target text for %s: '%s'", c.ID, targetText)
			} else {
				log.Info("Store Get %s %s returned empty/false", c.ID, convertTo)
			}
		}

		if found {
			log.Info("Comparing: Src='%s' Tgt='%s'", normalizedCurrent, utils.NormalizeCommentText(targetText))
		}

		if found {
			// Compare normalized texts to avoid unnecessary replacements
			normalizedTarget := utils.NormalizeCommentText(targetText)
			if normalizedTarget != normalizedCurrent {
				log.Info("Applying change: '%s' -> '%s'", normalizedCurrent, normalizedTarget)
				// Calculate offsets
				startLineIdx := c.Range.StartLine - 1
				endLineIdx := c.Range.EndLine - 1

				startOffset := lineOffsets[startLineIdx] + c.Range.StartCol - 1
				endOffset := lineOffsets[endLineIdx] + c.Range.EndCol - 1

				finalText := targetText
				// If targetText doesn't have markers, add them based on original type
				if c.Type == domain.CommentTypeLine && !strings.HasPrefix(targetText, "//") {
					finalText = "// " + targetText
				} else if c.Type == domain.CommentTypeBlock && !strings.HasPrefix(targetText, "/*") {
					finalText = "/* " + targetText + " */"
				}

				replacements = append(replacements, replacement{
					startOffset: startOffset,
					endOffset:   endOffset,
					newText:     finalText,
				})
			} else {
				log.Info("Skipping: text unchanged")
			}
		}
	}

	// Apply replacements in reverse order
	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].startOffset > replacements[j].startOffset
	})

	newSrc := string(src)
	for _, r := range replacements {
		// Safety check bounds
		if r.startOffset < 0 || r.endOffset > len(newSrc) {
			continue
		}
		newSrc = newSrc[:r.startOffset] + r.newText + newSrc[r.endOffset:]
	}

	if convertDryRun {
		fmt.Printf("File: %s\n", file)
		// Diff logic omitted for brevity
		return
	}

	if err := ioutil.WriteFile(file, []byte(newSrc), 0644); err != nil {
		log.Error("写入文件 %s 失败: %v", file, err)
	} else {
		log.Success("已处理 %s", file)
	}
}
