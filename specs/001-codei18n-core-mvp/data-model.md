# CodeI18n Core MVP 数据模型

本文档定义了 CodeI18n 系统的核心数据结构。所有结构体定义使用 Go 语言描述，并包含 JSON 序列化标签。

## 1. 核心实体

### Comment (注释)

表示从源代码中提取的单条注释信息。

```go
// TextRange 表示源码中的文本范围（1-based）
type TextRange struct {
    StartLine int `json:"startLine"` // 起始行号
    StartCol  int `json:"startCol"`  // 起始列号
    EndLine   int `json:"endLine"`   // 结束行号
    EndCol    int `json:"endCol"`    // 结束列号
}

// CommentType 定义注释类型
type CommentType string

const (
    CommentTypeLine  CommentType = "line"  // 单行注释 //
    CommentTypeBlock CommentType = "block" // 块注释 /* */
    CommentTypeDoc   CommentType = "doc"   // 文档注释
)

// Comment 表示从 AST 提取的注释
type Comment struct {
    // ID 是基于语义计算的唯一标识符
    // 计算规则: SHA1(file_path + language + symbol + normalized_text)
    ID string `json:"id"`

    // File 是相对于项目根目录的文件路径
    File string `json:"file"`

    // Language 是编程语言标识符 (如 "go", "rust")
    Language string `json:"language"`

    // Symbol 是注释绑定的语义符号路径 (如 "package.main.CalculateBalance")
    // 对于无法绑定到特定符号的注释，使用 "file.global" 或类似标识
    Symbol string `json:"symbol"`

    // Range 是注释在源码中的位置
    Range TextRange `json:"range"`

    // SourceText 是注释的原始文本内容 (英文)
    SourceText string `json:"sourceText"`

    // Type 是注释的类型
    Type CommentType `json:"type"`
}
```

**验证规则**:
- `ID` 必须是 40 字符的 SHA1 十六进制字符串。
- `File` 必须是有效的文件路径。
- `Language` 必须是支持的语言代码。
- `SourceText` 不能为空。

### LocalizedComment (本地化注释)

表示注释的多语言翻译内容。

```go
// LocalizedComment 表示注释的特定语言翻译
type LocalizedComment struct {
    // CommentID 引用对应的 Comment.ID
    CommentID string `json:"commentId"`

    // Lang 是目标语言代码 (遵循 BCP 47，如 "zh-CN", "ja")
    Lang string `json:"lang"`

    // Text 是翻译后的文本内容
    Text string `json:"text"`
}
```

**验证规则**:
- `CommentID` 必须存在于扫描到的注释列表中。
- `Lang` 必须符合 BCP 47 标准。

## 2. 存储实体

### Mapping (映射文件)

用于存储项目的所有多语言映射数据，通常保存为 `.codei18n/mappings.json`。

```go
// Mapping 存储整个项目的多语言映射
type Mapping struct {
    // Version 是映射文件格式的版本号 (如 "1.0")
    Version string `json:"version"`

    // SourceLanguage 是源码中的语言 (通常为 "en")
    SourceLanguage string `json:"sourceLanguage"`

    // TargetLanguage 是本地目标语言 (如 "zh-CN")
    TargetLanguage string `json:"targetLanguage"`

    // Comments 存储映射数据
    // 第一层 Key 是 Comment.ID
    // 第二层 Key 是语言代码 (如 "zh-CN")
    // Value 是翻译后的文本
    Comments map[string]map[string]string `json:"comments"`
}
```

**验证规则**:
- `Version` 必须匹配当前支持的版本。
- `Comments` 映射不能为空（初始化时除外）。

### Config (项目配置)

用于存储项目级配置，通常保存为 `.codei18n/config.json`。

```go
// Config 表示项目配置
type Config struct {
    // SourceLanguage 仓库源码中的注释语言 (默认 "en")
    SourceLanguage string `json:"sourceLanguage"`

    // LocalLanguage 本地开发显示的语言 (如 "zh-CN")
    LocalLanguage string `json:"localLanguage"`

    // ExcludePatterns 需要排除的文件或目录模式 (Glob 模式)
    ExcludePatterns []string `json:"excludePatterns"`

    // TranslationProvider 选用的翻译服务提供商 ("google", "openai")
    TranslationProvider string `json:"translationProvider"`

    // TranslationConfig 翻译服务的具体配置参数
    // 注意: 敏感信息如 API Key 不应存储在此处，应通过环境变量传递
    TranslationConfig map[string]string `json:"translationConfig"`
}
```

**验证规则**:
- `SourceLanguage` 和 `LocalLanguage` 不能为空。
- `TranslationProvider` 必须是支持的提供商之一。
