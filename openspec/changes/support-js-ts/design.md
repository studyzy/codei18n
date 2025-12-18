# 设计文档：JavaScript/TypeScript 适配器设计

## 架构概述

本设计复用 CodeI18n 现有的 Adapter 架构，通过 `adapters/javascript` 和 `adapters/typescript`（或统一为 `adapters/js_ts`）包实现 `LanguageAdapter` 接口。核心解析逻辑基于 `tree-sitter`。

### 核心组件

1.  **Tree-sitter Bindings**: 使用 `github.com/smacker/go-tree-sitter`，引入 `javascript` 和 `typescript/typescript`、`typescript/tsx` grammar。
2.  **Adapter 实现**:
    *   `NewAdapter()`: 初始化 parser，加载对应的 grammar。
    *   `Parse()`: 执行解析，遍历 AST 提取注释。
3.  **Symbol 解析 (关键)**:
    *   需要将注释绑定到具体的代码符号 (Symbol)。
    *   **策略**: 自底向上或自顶向下遍历，寻找注释节点紧邻的后续命名节点。
    *   **JS/TS 特有模式**:
        *   函数声明: `function foo() {}` -> `foo`
        *   类方法: `class A { method() {} }` -> `A.method`
        *   变量函数 (Arrow Function): `const bar = () => {}` -> `bar`
        *   对象属性方法: `const obj = { baz: function() {} }` -> `obj.baz`
        *   TypeScript 接口/类型: `interface I {}`, `type T = ...`
4.  **JSX/TSX 支持**:
    *   JSX 语法中的注释可能出现在 `{/* comment */}` 中，需要专门的 query 处理。

## 数据流

1.  **输入**: 源代码文件路径及内容。
2.  **解析**: Tree-sitter 生成 CST (Concrete Syntax Tree)。
3.  **查询**: 使用 Tree-sitter Query (S-expressions) 匹配注释节点及其上下文。
    *   例如：`(comment) @comment`
4.  **绑定**: 计算每个注释的 `Symbol` 路径。
5.  **输出**: `[]*domain.Comment` 列表，包含 ID、位置、原文、Symbol 等信息。

## 目录结构规划

建议在 `adapters/` 下创建 `typescript` 目录统一处理 JS/TS（因为 TS parser 通常兼容 JS）：

```
adapters/
  └── typescript/
      ├── adapter.go       # 实现 LanguageAdapter
      ├── queries.go       # 定义 tree-sitter queries
      ├── symbol.go        # 符号路径提取逻辑
      └── adapter_test.go  # 测试
```
或者，如果 `javascript` 和 `typescript` grammar 差异较大，可以分开或共享基础逻辑。考虑到 `tree-sitter-javascript` 和 `tree-sitter-typescript` 是不同的 parser，可能需要一个工厂模式内部区分，但对外暴露为一个 Adapter 或根据扩展名分派。

**决策**: 既然 Rust 也是独立目录，建议 `adapters/js_ts` 或 `adapters/typescript` (涵盖 JS)。鉴于 `codei18n` 目前按扩展名分派，可以在 `adapters/registry.go` 中将 `.js`, `.ts` 都指向同一个适配器实现，该实现内部根据文件后缀选择对应的 Tree-sitter Language。

## 依赖管理

需要 `go get` 相关的 tree-sitter bindings：
- `github.com/smacker/go-tree-sitter/javascript`
- `github.com/smacker/go-tree-sitter/typescript/typescript`
- `github.com/smacker/go-tree-sitter/typescript/tsx`

## 验证计划

1.  **单元测试**: 针对不同语法结构（Class, Function, Arrow Func, Interface, JSX）编写测试用例，验证 Symbol 提取的准确性。
2.  **集成测试**: 选取开源 JS/TS 项目文件进行端到端扫描测试。
