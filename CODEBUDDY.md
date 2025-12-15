# CODEBUDDY.md

本文件为 CodeBuddy Code 在此仓库中工作时提供指导。

**重要**: 本项目遵循位于 `.specify/memory/constitution.md` 的项目章程（v1.0.0）。所有开发工作必须符合章程中的核心原则和质量标准。

## 项目概述

CodeI18n 是面向工程团队的**代码注释国际化基础设施**。核心目标是实现：
- 源代码仓库维护统一语言（推荐英文）
- 本地开发环境以开发者母语显示注释
- Git 提交与 IDE 显示之间自动、可逆的翻译
- 零 Git 污染（无 diff/blame 污染）

这不是一个简单的翻译插件，而是围绕 AST、Git 和 IDE 集成构建的完整工具链。

## 开发命令

### 构建
```bash
go build -o codei18n ./cmd/codei18n
```

### 测试
```bash
# 运行所有测试
go test ./...

# 运行测试并生成覆盖率报告（必须 ≥ 60%，核心模块 ≥ 80%）
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 检查覆盖率百分比
go tool cover -func=coverage.out
```

### 代码检查
```bash
# 格式化代码（必需）
gofmt -w .

# 静态检查（必需）
golint ./...
# 或使用
staticcheck ./...
```

### 运行
```bash
# CLI 用于扫描和翻译
./codei18n scan --file xxx.go --lang zh-CN --format json
```

## 架构总览

### 系统分层

项目分为三个主要层次：

1. **CodeI18n Core (Go)** - 位于 `core/` 目录
   - 支持不同编程语言的 AST 解析器
   - 注释模型和统一抽象
   - 注释 ID 生成（稳定、语义绑定）
   - 多语言注释映射存储
   - 翻译引擎

2. **语言适配器** - 位于 `adapters/` 目录
   - Go 适配器：使用 `go/parser` 和 `go/ast`
   - Rust 适配器：使用 `syn`（通过 sidecar）或 tree-sitter
   - Tree-sitter 适配器：用于 JS/TS/Java/Python/C#（规划中）

3. **IDE 插件** - 位于 `ide/` 目录
   - VS Code 扩展：使用 TextEditorDecorationType 进行渲染
   - JetBrains 插件：使用 InlayModel/Inlay Hints

### 目录结构

```
CodeI18n/
├── cmd/codei18n         # CLI 入口
├── core/                # 核心基础设施
│   ├── comment/         # 注释模型和解析器
│   ├── mapping/         # 多语言映射存储
│   └── translate/       # 翻译引擎
├── adapters/            # 编程语言适配器
│   ├── go/             # Go AST 适配器
│   ├── rust/           # Rust 适配器（规划中）
│   └── treesitter/     # 基于 Tree-sitter 的适配器（规划中）
└── ide/                # IDE 集成
    ├── vscode/         # VS Code 扩展
    └── jetbrains/      # JetBrains 插件
```

## 核心设计原则

### 1. 单一事实源（Single Source of Truth）
- 源码中的英文注释 = 唯一事实源
- 本地语言注释 = 派生视图
- 绝不在 Git 中存储多个语言版本

### 2. 语义优先，而非文本优先
- 注释必须绑定到：
  - 文件路径
  - 语义符号（函数/结构体/模块）
- **禁止**：基于行号或正则表达式处理注释

### 3. Core 与 IDE 解耦
- **Core (Go)**：注释解析、ID 生成、翻译、映射
- **IDE 插件**：仅负责渲染，不包含翻译/AST 逻辑

## 注释模型

### 统一注释抽象
```go
type Comment struct {
    ID          string      // 稳定的 SHA1 哈希
    File        string      // 文件路径
    Language    string      // go/rust/js/...
    Symbol      string      // func/struct/impl/module
    Range       TextRange   // 源码位置
    SourceText  string      // 英文文本
}

type LocalizedComment struct {
    CommentID string        // 引用 Comment.ID
    Lang      string        // zh-CN/en/ja/...
    Text      string        // 翻译后的文本
}
```

### 注释 ID 生成

**关键设计**：注释 ID 必须：
- 在代码重构时保持稳定
- 绑定到代码语义，而非行号
- 确定性生成

**ID 计算方式**：
```
ID = SHA1(
    file_path +
    language +
    parent_symbol +
    normalized_comment_text
)
```

**parent_symbol 示例**：
- Go: `package.func`
- Rust: `impl::fn`
- Java: `Class#method`

## 语言适配器接口

所有编程语言支持都遵循以下契约：

```go
type LanguageAdapter interface {
    Language() string
    Parse(file string) ([]Comment, error)
}
```

### Go 适配器（第一阶段）
- 使用 `go/parser` 和 `go/ast`
- 支持：`//`、`/* */`、文档注释
- 注释天然绑定到 AST 节点

### Rust 适配器（第二阶段）
- 通过 `syn` crate（sidecar 进程）或 tree-sitter-rust 获取 AST
- 支持：`//`、`///`、`/** */`

### Tree-sitter 适配器（第三阶段）
- 统一方式支持 JS/TS/Java/Python/C#

## 翻译引擎

### 接口
```go
type Translator interface {
    Translate(text, from, to string) (string, error)
}
```

### 策略
- **本地缓存优先**（避免提交时实时调用 API）
- 支持的后端：
  - DeepL
  - OpenAI
  - 自定义术语词典
- 翻译失败可阻断提交（可配置）

## 映射文件

### 存储位置
- 默认路径：`.codei18n/`
- 映射文件仅本地使用（默认不提交到 Git）

### 文件格式
```json
{
  "version": "1.0",
  "sourceLanguage": "en",
  "targetLanguage": "zh-CN",
  "comments": {
    "a8f9c3e2": {
      "en": "Calculate account balance",
      "zh-CN": "计算账户余额"
    }
  }
}
```

## Git 工作流

### pre-commit Hook 行为
```
git commit
 └─ pre-commit
    ├─ 扫描 staged 文件
    ├─ 检测非英文注释
    ├─ 翻译为英文
    ├─ 更新映射文件
    └─ 替换源码注释
```

**关键原则**：
- 仅处理 staged 文件
- 翻译失败可阻断提交（可配置）
- 保持可逆性

## IDE 集成

### VS Code 扩展
- 使用带 `after` 内容的 `TextEditorDecorationType`
- 覆盖/附加母语注释显示
- **不**修改文档内容

### JetBrains 插件
- 使用 Kotlin 实现（IntelliJ Platform Plugin）
- 使用 `InlayModel` 实现 Inlay Hints
- 显示格式：`// English comment` 后跟 `// ↓ 本地语言注释`
- 备选方案：Gutter 图标 + hover tooltip

### CLI 输出格式
```bash
codei18n scan --file xxx.go --lang zh-CN --format json
```
返回结构化 JSON 供 IDE 使用。

## 配置文件

位置：`.codei18n/config.json`

```json
{
  "sourceLanguage": "en",
  "localLanguage": "zh-CN",
  "ide": {
    "vscode": {
      "displayMode": "overlay"
    },
    "jetbrains": {
      "displayMode": "inlay"
    }
  }
}
```

## 开发路线图

### v0.1 (MVP)
- Go AST 适配器
- 用于注释扫描和翻译的 CLI
- VS Code 扩展

### v0.2
- Rust 适配器
- pre-commit Git hook
- JetBrains 插件（GoLand/RustRover）

### v0.3
- Tree-sitter 适配器
- JS/TS/Java 支持

## 关键约束（不可协商）

**来自项目章程 v1.0.0**:

1. **AST 优先**: 所有注释解析**必须**基于 AST/PSI（绝不使用基于行号或正则表达式的方式）
2. **单一语言源**: Git 仓库**必须**只包含一种注释语言（英文）
3. **代码规范**: **必须**遵循 Effective Go，使用 gofmt 和 golint/staticcheck
4. **测试覆盖率**: **必须**达到 60% 以上（核心模块 comment/mapping/translate 必须 80% 以上）
5. **中文优先**: 所有文档和代码注释**必须**使用中文
6. **CLI 优先**: Core 和 IDE 插件**必须**解耦，核心功能通过 CLI 暴露
7. **注释翻译**: **必须**可逆且可定位

## 重要说明

- 这是一个全新项目 - 实现基于 README.md 中的技术设计
- 项目使用 Go 1.25+（参见 go.mod）
- 基于 AST 的方法至关重要 - 绝不使用正则表达式或基于行号的注释提取
- `.specify/` 目录包含项目规范工具（SpecKit 工作流）
- **项目章程**: 所有开发决策必须符合 `.specify/memory/constitution.md` 中定义的原则
- **IDE 插件**: VS Code 和 JetBrains 插件将在独立仓库实现，本仓库专注于 Core 引擎
