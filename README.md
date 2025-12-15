# CodeI18n 技术文档

**Code Comment Internationalization Infrastructure**

---

## 1. 项目定位与问题定义

### 1.1 项目定位

**CodeI18n** 是一个面向工程团队的 **代码注释国际化基础设施**，目标是在不破坏 Git 语义、不污染源码、不影响编译与调试的前提下，实现：

* **源码仓库统一语言（推荐英文）**
* **本地开发环境以母语阅读和编写注释**
* **注释语言在 Git 提交与 IDE 展示之间自动、可逆转换**

CodeI18n 不是简单的“翻译插件”，而是一个 **围绕 AST、Git、IDE 的完整工具链**。

---

### 1.2 要解决的核心问题

| 问题              | 传统方案      | CodeI18n      |
| --------------- | --------- | ------------- |
| 注释语言冲突          | 强制英文 / 中文 | 英文入库，本地多语言    |
| diff / blame 污染 | 严重        | 不污染           |
| 多语言协作           | 成本高       | 零成本           |
| IDE 体验          | 无         | 原生集成          |
| 可扩展性            | 无         | AST + Adapter |

---

## 2. 总体设计目标

### 2.1 功能目标

* 支持 **多自然语言**

    * 默认：英文（en）、中文（zh-CN）
    * 可扩展：ja、ko、fr 等
* 支持 **多编程语言**

    * 第一阶段：Go
    * 第二阶段：Rust
    * 第三阶段：JS / TS / Java / Python / C#
* 支持 **多 IDE**

    * VS Code
    * JetBrains 家族（IntelliJ / GoLand / RustRover 等）

---

### 2.2 非功能目标（硬约束）

1. **Git 仓库中只存在一种注释语言**
2. **IDE 展示不修改源码**
3. **所有注释解析必须基于 AST / PSI**
4. **注释翻译必须可逆、可定位**
5. **Core 与 IDE 插件解耦**

---

## 3. 核心设计原则

### 3.1 单一事实源（Single Source of Truth）

* 英文注释源码 = 唯一事实源
* 本地语言注释 = 派生视图

---

### 3.2 语义优先，而非文本优先

* 注释必须绑定到：

    * 文件
    * 语义符号（函数 / 结构体 / 模块）
* 禁止基于行号、正则处理注释

---

### 3.3 Core 与 IDE 解耦

* **Core（Go 实现）**

    * 注释解析
    * ID 生成
    * 翻译与映射
* **IDE 插件**

    * 只负责渲染
    * 不包含翻译、AST 逻辑

---

## 4. 整体系统架构

### 4.1 架构总览

```text
┌────────────────────────────────────────────┐
│                   IDE Layer                │
│                                            │
│  ┌───────────────┐   ┌──────────────────┐ │
│  │ VS Code       │   │ JetBrains IDEs   │ │
│  │ Extension     │   │ Plugin           │ │
│  └──────▲────────┘   └────────▲─────────┘ │
│         │ Decorations / Inlay           │   │
└─────────┼────────────────────────────────┘
          │ JSON / CLI
┌─────────┴────────────────────────────────┐
│          CodeI18n Core (Go)               │
│  - AST Parsers                            │
│  - Comment Model                          │
│  - Comment ID Generator                   │
│  - Mapping Store                          │
│  - Translation Engine                     │
└─────────▲────────────────────────────────┘
          │ pre-commit
┌─────────┴────────────────────────────────┐
│          Git Hooks / CLI                  │
└─────────▲────────────────────────────────┘
          │
┌─────────┴────────────────────────────────┐
│          Git Remote Repository            │
│        (English-only Source Code)         │
└───────────────────────────────────────────┘
```

---

## 5. CodeI18n Core 设计

### 5.1 Core 职责

* 扫描源码并提取注释
* 为注释生成稳定 ID
* 管理多语言注释映射
* 执行注释语言转换（提交前）
* 为 IDE 提供结构化注释数据

---

### 5.2 注释统一抽象模型

```go
type Comment struct {
    ID          string
    File        string
    Language    string          // go / rust / ...
    Symbol      string          // func / struct / impl / module
    Range       TextRange
    SourceText  string          // 英文
}
```

```go
type LocalizedComment struct {
    CommentID string
    Lang      string            // zh-CN / en / ...
    Text      string
}
```

---

## 6. 注释 ID 设计（关键）

### 6.1 ID 生成原则

注释 ID 必须满足：

* 稳定
* 与代码语义绑定
* 不依赖行号

### 6.2 ID 计算方式（推荐）

```text
ID = SHA1(
    file_path +
    language +
    parent_symbol +
    normalized_comment_text
)
```

📌 **parent_symbol 示例**：

* Go：`package.func`
* Rust：`impl::fn`
* Java：`Class#method`

---

## 7. 多自然语言支持设计

### 7.1 映射文件结构

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

### 7.2 存储策略

* 默认路径：`.codei18n/`
* 映射文件：

    * 本地使用
    * **默认不提交 Git**

---

## 8. 多编程语言支持策略

### 8.1 编程语言适配器接口

```go
type LanguageAdapter interface {
    Language() string
    Parse(file string) ([]Comment, error)
}
```

---

### 8.2 Go 语言实现（第一阶段）

* 使用：

    * `go/parser`
    * `go/ast`
* 支持：

    * `//`
    * `/* */`
    * Doc comment

优势：

* AST 成熟
* 注释天然绑定节点

---

### 8.3 Rust 语言实现（第二阶段）

* AST 技术路线：

    * `syn`（通过 sidecar）
    * 或 tree-sitter-rust
* 处理：

    * `//`
    * `///`
    * `/** */`

---

### 8.4 其他语言（第三阶段）

统一走 **Tree-sitter Adapter**：

| 语言      | 状态 |
| ------- | -- |
| JS / TS | 计划 |
| Java    | 计划 |
| Python  | 计划 |
| C#      | 计划 |

---

## 9. IDE 支持设计

### 9.1 IDE 统一交互模式

IDE 插件通过调用 CLI 获取结构化数据：

```bash
codei18n scan --file xxx.go --lang zh-CN --format json
```

---

## 10. VS Code 插件设计

### 10.1 渲染方式

* `TextEditorDecorationType`
* 覆盖 / 附加显示中文注释
* 不修改文档内容

```ts
after: {
  contentText: "计算账户余额"
}
```

---

## 11. JetBrains IDE 插件设计

### 11.1 插件类型

* IntelliJ Platform Plugin
* Kotlin 实现
* 多 IDE 共用

---

### 11.2 渲染方案（推荐）

#### Inlay Hint

```text
// Calculate account balance
// ↓ 计算账户余额
```

* 使用 `InlayModel`
* 官方推荐方式
* 不影响 PSI / LSP

---

### 11.3 备选方案

* Gutter icon + hover tooltip

---

## 12. Git 提交流程设计

### 12.1 pre-commit Hook

```text
git commit
 └─ pre-commit
    ├─ 扫描 staged 文件
    ├─ 识别非英文注释
    ├─ 翻译为英文
    ├─ 更新映射文件
    └─ 替换源码注释
```

### 12.2 行为原则

* 只处理 staged 文件
* 翻译失败可阻断提交（可配置）

---

## 13. 翻译引擎设计

### 13.1 抽象接口

```go
type Translator interface {
    Translate(text, from, to string) (string, error)
}
```

### 13.2 实现策略

* 本地缓存优先
* 支持：

    * DeepL
    * OpenAI
    * 术语表
* 避免提交时实时调用大模型

---

## 14. 配置文件设计

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

---

## 15. 项目目录结构建议

```text
CodeI18n/
├─ cmd/codei18n
├─ core/
│  ├─ comment/
│  ├─ mapping/
│  ├─ translate/
├─ adapters/
│  ├─ go/
│  ├─ rust/
│  └─ treesitter/
├─ ide/
│  ├─ vscode/
│  └─ jetbrains/
├─ docs/
└─ examples/
```

---

## 16. Roadmap（明确可执行）

### v0.1

* Go AST
* CLI
* VS Code

### v0.2

* Rust
* pre-commit
* JetBrains（GoLand / RustRover）

### v0.3

* Tree-sitter
* JS / TS / Java

---

## 17. 结论

**CodeI18n 本质是一个“代码注释 i18n 基础设施”**：

* Git 负责事实
* AST 负责语义
* IDE 负责体验

这是一个 **长期存在、但一直没人系统解决的问题**。
