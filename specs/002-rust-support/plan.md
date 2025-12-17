# 技术实现计划: Rust 语言支持

**功能**: Rust 语言支持 (`rust-support`)
**相关规范**: [specs/002-rust-support/spec.md](spec.md)
**创建时间**: 2025-12-17
**状态**: 规划中

## 架构设计

### 核心组件

1.  **RustAdapter (`adapters/rust/adapter.go`)**
    *   实现 `core/interfaces.LanguageAdapter` 接口。
    *   负责加载 Rust 代码，调用 Tree-sitter 解析，提取注释，生成 ID。
    *   **关键依赖**: `github.com/smacker/go-tree-sitter` 和 `github.com/smacker/go-tree-sitter/rust`。

2.  **Tree-sitter 解析器集成**
    *   使用 CGO 静态链接 Rust 语法库。
    *   利用 S-expression 查询 (`Query`) 提取注释节点。

3.  **注释处理器 (`adapters/rust/extractor.go`)**
    *   **注释分类**: 区分 `//` (普通), `///` (文档), `//!` (模块文档) 等。
    *   **上下文解析**:
        *   对于 `///` 和 `/**`: 向下查找最近的非属性（Attribute）节点作为宿主符号（Owner Symbol）。
        *   对于 `//!` 和 `/*!`: 以父节点（通常是 Module 或 Root）作为宿主符号。
        *   对于 `//`: 向上查找最近的“具名符号”（Function, Struct, Mod 等）构建上下文路径。
    *   **ID 生成**: `MD5(FilePath + SymbolPath + CommentContent)`。

4.  **符号路径构建器 (`adapters/rust/symbol.go`)**
    *   遍历 AST 节点路径，构建类似 `my_mod::MyStruct::my_fn` 的语义路径。
    *   支持的节点类型: `source_file`, `mod_item`, `impl_item`, `struct_item`, `enum_item`, `function_item`, `trait_item`。

### 数据流

1.  **提取 (Extract)**: `RustAdapter.Extract(path)` -> 读取文件 -> Tree-sitter Parse -> Run Query `(line_comment) @c (block_comment) @c` -> 遍历结果 -> 解析符号上下文 -> 返回 `[]*domain.Comment`。
2.  **注入 (Inject)**: `RustAdapter.Inject(path, comments)` -> 读取文件 -> 替换注释内容 (基于原始文本定位或 AST 范围) -> 写回文件。 *注: 初始版本可简化为基于文本替换，只要 ID 匹配。*

## 任务分解

### Phase 1: 基础设施搭建
- [ ] **Task 1.1**: 添加 `smacker/go-tree-sitter` 依赖并在 `go.mod` 中确认。
- [ ] **Task 1.2**: 创建 `adapters/rust` 包结构，实现基本的 `LanguageAdapter` 存根（Stub）。
- [ ] **Task 1.3**: 编写简单的 `TestRustParser` 验证 Tree-sitter 能正确解析 Rust 代码（Hello World）。

### Phase 2: 注释提取逻辑
- [ ] **Task 2.1**: 实现 `Query` 逻辑，提取所有 `line_comment` 和 `block_comment` 节点。
- [ ] **Task 2.2**: 实现“注释类型识别”逻辑（区分 Doc vs Normal）。
- [ ] **Task 2.3**: 实现“上下文宿主解析”逻辑（NextSibling 处理，跳过 Attributes）。
- [ ] **Task 2.4**: 实现“符号路径构建”逻辑（Ancestor 遍历）。
- [ ] **Task 2.5**: 集成所有逻辑到 `Extract` 方法，并编写单元测试验证 JSON 输出。

### Phase 3: ID 生成与稳定性
- [ ] **Task 3.1**: 实现 Rust 特定的 ID 生成策略。
- [ ] **Task 3.2**: 编写“抗格式化干扰”测试：验证格式化前后 ID 一致性。

### Phase 4: 注入与集成
- [ ] **Task 4.1**: 实现 `Inject` 方法（基于 AST Range 或简单的字符串替换）。
- [ ] **Task 4.2**: 在 `core/handler` 或 `main.go` 中注册 `RustAdapter`。
- [ ] **Task 4.3**: 端到端测试：扫描 Rust 项目 -> 翻译(Mock) -> 应用翻译。

## 技术细节

### Tree-sitter Query
```scm
(line_comment) @comment
(block_comment) @comment
```
*注：由于需要复杂的上下文判断（如跳过 Attributes），查询仅用于获取所有注释，后续逻辑在 Go 中处理更灵活。*

### 符号节点映射
需要关注的 Tree-sitter 节点类型：
*   `function_item` -> name: `identifier`
*   `struct_item` -> name: `type_identifier`
*   `mod_item` -> name: `identifier`
*   `impl_item` -> type: `type_identifier` (注意处理 trait impl `impl A for B`)

## 风险与缓解

*   **Risk**: CGO 编译问题（特别是跨平台）。
    *   *Mitigation*: 在 CI 中配置多平台构建测试。开发文档明确 CGO 要求。
*   **Risk**: Tree-sitter 节点类型变更。
    *   *Mitigation*: 锁定 `go-tree-sitter` 和 grammar 版本。
*   **Risk**: 宏（Macros）导致 AST 结构混乱。
    *   *Mitigation*: 对于无法识别的宏内部注释，降级为“文件级”或“未知上下文”，保证不崩溃。

## 验证计划

1.  **单元测试**: 覆盖提取、ID 生成、路径构建。
2.  **集成测试**: 使用真实 Rust 开源项目文件（如 `actix-web` 或 `tokio` 的片段）进行扫描测试。
