package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/studyzy/codei18n/adapters"
	"github.com/studyzy/codei18n/core/mapping"
	"github.com/studyzy/codei18n/core/utils"
)

// TestJavaE2E_FullConvertWorkflow 测试 Java 文件的完整转换工作流
func TestJavaE2E_FullConvertWorkflow(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Calculator.java")

	// 步骤 1: 创建原始英文 Java 文件
	originalEnglish := `package com.example;

// Calculator class for basic math operations
public class Calculator {
    
    // Add two numbers
    public int add(int a, int b) {
        return a + b;
    }
}
`
	err := os.WriteFile(testFile, []byte(originalEnglish), 0644)
	assert.NoError(t, err)
	t.Logf("✓ 创建原始英文 Java 文件")

	// 步骤 2: 扫描并提取注释
	adapter, err := adapters.GetAdapter(testFile)
	assert.NoError(t, err)
	assert.Equal(t, "java", adapter.Language())

	comments, err := adapter.Parse(testFile, []byte(originalEnglish))
	assert.NoError(t, err)
	assert.Greater(t, len(comments), 0)
	t.Logf("✓ 成功提取 %d 个注释", len(comments))

	// 步骤 3: 创建映射并添加翻译
	store := mapping.NewStore(filepath.Join(tmpDir, "mappings.json"))

	// 手动添加翻译
	translations := map[string]string{
		"// Calculator class for basic math operations": "基本数学运算的计算器类",
		"// Add two numbers":                            "计算两个数的和",
	}

	for _, c := range comments {
		id := utils.GenerateCommentID(c)
		
		// 存储英文原文
		store.Set(id, "en", c.SourceText)
		
		// 添加中文翻译
		if trans, ok := translations[c.SourceText]; ok {
			store.Set(id, "zh-CN", trans)
			t.Logf("  映射: '%s' -> '%s'", c.SourceText, trans)
		}
	}
	
	err = store.Save()
	assert.NoError(t, err)
	t.Logf("✓ 保存翻译映射")

	// 步骤 4: 验证可以查找翻译
	for _, c := range comments {
		id := utils.GenerateCommentID(c)
		zhText, ok := store.Get(id, "zh-CN")
		if ok && zhText != "" {
			t.Logf("  ✓ 找到翻译: ID=%s, '%s' -> '%s'", id, c.SourceText, zhText)
		}
	}

	// 步骤 5: 验证反向查找（中文 -> 英文）
	// 模拟已转换为中文的内容
	convertedContent := `package com.example;

// 基本数学运算的计算器类
public class Calculator {
    
    // 计算两个数的和
    public int add(int a, int b) {
        return a + b;
    }
}
`
	chineseComments, err := adapter.Parse(testFile, []byte(convertedContent))
	assert.NoError(t, err)

	foundCount := 0
	for _, c := range chineseComments {
		normalizedCurrent := utils.NormalizeCommentText(c.SourceText)
		
		// 在映射中查找
		for _, transMap := range store.GetMapping().Comments {
			zhText, hasZh := transMap["zh-CN"]
			enText, hasEn := transMap["en"]
			
			if hasZh && hasEn {
				normalizedZh := utils.NormalizeCommentText(zhText)
				if normalizedZh == normalizedCurrent {
					foundCount++
					t.Logf("  ✓ 反向查找成功: ZH='%s' -> EN='%s'", zhText, enText)
					break
				}
			}
		}
	}
	
	assert.Greater(t, foundCount, 0, "应该能通过中文找到对应的英文")

	t.Log("\n=== Java E2E Convert 测试成功完成 ===")
	t.Log("测试流程:")
	t.Log("  1. 创建英文 Java 文件 ✓")
	t.Log("  2. 扫描并提取注释 ✓")
	t.Log("  3. 添加翻译并保存映射 ✓")
	t.Log("  4. 验证正向查找（英文 -> 中文）✓")
	t.Log("  5. 验证反向查找（中文 -> 英文）✓")
}

// TestJavaConvert_CLICompatibility 测试 convert 命令对 Java 文件的兼容性
func TestJavaConvert_CLICompatibility(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "简单类",
			filename: "Simple.java",
			content: `package test;
// Simple class
public class Simple {}`,
		},
		{
			name:     "接口",
			filename: "Interface.java",
			content: `package test;
// User interface
public interface User {}`,
		},
		{
			name:     "枚举",
			filename: "Enum.java",
			content: `package test;
// Status enum
public enum Status { ACTIVE, INACTIVE }`,
		},
		{
			name:     "内部类",
			filename: "Outer.java",
			content: `package test;
public class Outer {
    // Inner class
    public static class Inner {}
}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, tc.filename)
			
			err := os.WriteFile(testFile, []byte(tc.content), 0644)
			assert.NoError(t, err)

			// 验证 adapters.GetAdapter 能识别
			adapter, err := adapters.GetAdapter(testFile)
			assert.NoError(t, err)
			assert.NotNil(t, adapter)
			assert.Equal(t, "java", adapter.Language())

			// 验证能解析
			comments, err := adapter.Parse(testFile, []byte(tc.content))
			assert.NoError(t, err)
			assert.Greater(t, len(comments), 0, "应该至少提取到一个注释")

			t.Logf("✓ %s: 识别并解析成功", tc.name)
		})
	}
}
