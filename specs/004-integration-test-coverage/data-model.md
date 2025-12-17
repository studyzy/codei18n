# 数据模型: Integration Test Coverage

**功能**: Integration Test Coverage
**状态**: 草稿
**日期**: 2025-12-17

## 1. 核心实体

### 1.1 CLI 输出模型 (JSON)

这是 CLI `scan` 命令在 `--format json` 模式下的输出数据模型。这也是 IDE 插件所依赖的"协议"。

#### 实体: `ScanResult` (单文件)

当扫描单个文件时 (`--file`) 的输出结构。

| 字段 | 类型 | 描述 |
|---|---|---|
| `file` | `string` | 扫描的文件路径 (可选) |
| `comments` | `Comment[]` | 提取出的注释列表 |

#### 实体: `Comment`

代表源代码中的一个注释单元。

| 字段 | 类型 | 描述 | 示例 |
|---|---|---|---|
| `id` | `string` | 唯一标识符 (SHA Hash) | `"a1b2c3..."` |
| `file` | `string` | 所属文件路径 | `"main.go"` |
| `language` | `string` | 编程语言 | `"go"` |
| `symbol` | `string` | 关联的语法符号路径 | `"main.Func"` |
| `range` | `TextRange` | 代码位置范围 | `{...}` |
| `sourceText` | `string` | 原始注释文本 | `"Hello world"` |
| `type` | `enum` | 注释类型 | `"line"`, `"block"`, `"doc"` |
| `localizedText` | `string` | (可选) 翻译后的文本 | `"你好世界"` |

#### 实体: `TextRange`

| 字段 | 类型 | 描述 |
|---|---|---|
| `startLine` | `int` | 起始行号 (1-based) |
| `startCol` | `int` | 起始列号 (1-based) |
| `endLine` | `int` | 结束行号 (1-based) |
| `endCol` | `int` | 结束列号 (1-based) |

## 2. 测试夹具 (Fixtures)

用于集成测试的预定义数据。

### 2.1 简单 Go 文件 (fixtures/simple.go)
包含基本的行注释和块注释。

```go
package main

// Hello World
func main() {
    /* Block Comment */
}
```

### 2.2 复杂 Go 文件 (fixtures/complex.go)
包含文档注释、连续的多行注释 (`// ...`)、行内混合注释和被注释掉的代码。

```go
package main

// Line 1
// Line 2
func complex() {
    code() // Inline comment
    /* 
       Block 
       Comment 
    */
    // commented_out_code();
}
```

### 2.3 Rust 文件 (fixtures/lib.rs)
包含 Rust 特有的文档注释。

```rust
//! Inner doc comment for module

/// Outer doc comment for function
fn main() {
    // Line comment
    /* Block comment */
}
```

### 2.4 配置文件 (fixtures/config.json)
用于测试配置加载优先级。

```json
{
  "sourceLanguage": "en",
  "targetLanguage": "ja",
  "exclude": ["vendor/**"]
}
```
