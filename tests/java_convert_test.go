package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/studyzy/codei18n/adapters"
	"github.com/studyzy/codei18n/core/mapping"
	"github.com/studyzy/codei18n/core/utils"
)

func TestJavaConvert_ApplyMode(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Test.java")

	originalContent := `package com.example;

// Calculate sum
public class Calculator {
    // Add two numbers
    public int add(int a, int b) {
        return a + b;
    }
}
`
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	assert.NoError(t, err)

	// 获取适配器
	adapter, err := adapters.GetAdapter(testFile)
	assert.NoError(t, err)
	assert.Equal(t, "java", adapter.Language())

	// 解析注释
	comments, err := adapter.Parse(testFile, []byte(originalContent))
	assert.NoError(t, err)
	assert.Greater(t, len(comments), 0)

	// 创建映射存储并添加翻译
	store := mapping.NewStore("")
	
	expectedTranslations := map[string]string{
		"// Calculate sum":    "计算总和",
		"// Add two numbers":  "计算两个数的和",
	}

	// 为每个注释生成 ID 并添加翻译
	for _, c := range comments {
		id := utils.GenerateCommentID(c)
		if trans, ok := expectedTranslations[c.SourceText]; ok {
			store.Set(id, "en", c.SourceText)
			store.Set(id, "zh-CN", trans)
			t.Logf("设置翻译: ID=%s, EN='%s', ZH='%s'", id, c.SourceText, trans)
		}
	}

	// 验证可以找到翻译
	foundCount := 0
	for _, c := range comments {
		id := utils.GenerateCommentID(c)
		zhText, ok := store.Get(id, "zh-CN")
		if ok && zhText != "" {
			foundCount++
			t.Logf("找到翻译: '%s' -> '%s'", c.SourceText, zhText)
		}
	}

	assert.Greater(t, foundCount, 0, "应该至少找到一个翻译")
}

func TestJavaConvert_RestoreMode(t *testing.T) {
	// 创建包含中文注释的 Java 文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Test.java")

	// 已经转换为中文的内容
	convertedContent := `package com.example;

// 计算总和
public class Calculator {
    // 计算两个数的和
    public int add(int a, int b) {
        return a + b;
    }
}
`
	err := os.WriteFile(testFile, []byte(convertedContent), 0644)
	assert.NoError(t, err)

	// 获取适配器
	adapter, err := adapters.GetAdapter(testFile)
	assert.NoError(t, err)

	// 解析注释（当前是中文）
	comments, err := adapter.Parse(testFile, []byte(convertedContent))
	assert.NoError(t, err)

	// 创建映射存储，模拟已有的双语映射
	store := mapping.NewStore("")
	
	// 构建原始英文版本的注释来生成正确的 ID
	originalEnglishContent := `package com.example;

// Calculate sum
public class Calculator {
    // Add two numbers
    public int add(int a, int b) {
        return a + b;
    }
}
`
	
	// 解析英文版本以获取正确的 ID
	englishComments, err := adapter.Parse(testFile, []byte(originalEnglishContent))
	assert.NoError(t, err)

	translations := map[string]string{
		"// Calculate sum":   "计算总和",
		"// Add two numbers": "计算两个数的和",
	}

	// 使用英文注释的 ID 存储翻译
	for _, c := range englishComments {
		id := utils.GenerateCommentID(c)
		if trans, ok := translations[c.SourceText]; ok {
			store.Set(id, "en", c.SourceText)
			store.Set(id, "zh-CN", trans)
		}
	}

	// 现在验证可以通过中文文本反向查找英文
	for _, c := range comments {
		// 规范化当前注释文本
		normalizedCurrent := utils.NormalizeCommentText(c.SourceText)
		
		// 在映射中查找匹配的中文文本
		found := false
		for id, transMap := range store.GetMapping().Comments {
			zhText, hasZh := transMap["zh-CN"]
			enText, hasEn := transMap["en"]
			
			if hasZh && hasEn {
				normalizedZh := utils.NormalizeCommentText(zhText)
				if normalizedZh == normalizedCurrent {
					found = true
					t.Logf("反向查找成功: ZH='%s' -> EN='%s' (ID=%s)", zhText, enText, id)
					break
				}
			}
		}
		
		assert.True(t, found, "应该能通过中文文本找到对应的英文翻译: %s", c.SourceText)
	}
}

func TestJavaConvert_FileSupport(t *testing.T) {
	// 测试 convert 命令能识别 .java 文件
	tmpDir := t.TempDir()
	
	testFiles := []string{
		"Test.java",
		"Sample.java",
		"Application.java",
	}
	
	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		content := "package test;\n// Comment\npublic class Test {}"
		err := os.WriteFile(path, []byte(content), 0644)
		assert.NoError(t, err)
		
		// 验证能获取适配器
		adapter, err := adapters.GetAdapter(path)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "java", adapter.Language())
	}
}

func TestJavaConvert_CommentMarkers(t *testing.T) {
	// 测试不同类型的注释标记是否正确处理
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Test.java")

	content := `package com.example;

// Line comment
public class Test {
    /* Block comment */
    int x;
    
    /**
     * Javadoc comment
     */
    void method() {}
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	assert.NoError(t, err)

	adapter, err := adapters.GetAdapter(testFile)
	assert.NoError(t, err)

	comments, err := adapter.Parse(testFile, []byte(content))
	assert.NoError(t, err)

	// 验证注释类型和标记
	for _, c := range comments {
		if c.Type == "line" {
			assert.True(t, strings.HasPrefix(c.SourceText, "//"), 
				"行注释应该以 // 开头: %s", c.SourceText)
		} else if c.Type == "block" {
			assert.True(t, strings.HasPrefix(c.SourceText, "/*") && 
				strings.HasSuffix(c.SourceText, "*/"),
				"块注释应该以 /* 开头和 */ 结尾: %s", c.SourceText)
		} else if c.Type == "doc" {
			assert.True(t, strings.HasPrefix(c.SourceText, "/**"),
				"文档注释应该以 /** 开头: %s", c.SourceText)
		}
	}
}
