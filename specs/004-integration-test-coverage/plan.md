# 实施计划: Integration Test Coverage

**分支**: `004-integration-test-coverage` | **日期**: 2025-12-17 | **规范**: [spec.md](./spec.md)
**输入**: 来自 `/specs/004-integration-test-coverage/spec.md` 的功能规范

**注意**: 此模板由 `/speckit.plan` 命令填充. 执行工作流程请参见 `.specify/templates/commands/plan.md`.

## 摘要

本计划旨在建立完善的集成测试体系，重点覆盖 CLI 的核心功能和 IDE 插件交互协议。我们将扩展现有的 `tests/` 目录，使用 Go 标准库 `testing` 和 `os/exec` 模拟真实的 CLI 调用场景，验证 JSON 输出格式（IDE 协议）、参数解析优先级和错误处理机制。此外，我们将专门针对复杂场景进行测试，包括 Go 语言的各种注释变体（连续多行、行内、注释掉的代码）以及 Rust 语言的多语言支持验证。

## 技术背景

<!--
  需要操作: 将此部分内容替换为项目的技术细节.
  此处的结构以咨询性质呈现, 用于指导迭代过程.
-->

**语言/版本**: Go 1.25.5
**主要依赖**: Cobra (CLI), Viper (Config), Testify (Assert), OpenAI (Mocked)
**存储**: 文件系统 (JSON 映射文件)
**测试**: standard `testing` package + `github.com/stretchr/testify`
**目标平台**: Cross-platform (macOS/Linux/Windows) via Go runtime
**项目类型**: CLI Tool + Library
**性能目标**: 测试套件运行时间 < 30s
**约束条件**: 必须支持离线测试 (Mock LLM)
**规模/范围**: 扩展现有测试套件，增加约 15-20 个关键场景，覆盖多语言和边缘情况

## 章程检查

*门控: 必须在阶段 0 研究前通过. 阶段 1 设计后重新检查. *

**I. AST 优先原则**:
- [x] 所有注释处理使用 AST（go/parser、go/ast、syn、tree-sitter） - *测试将验证这一点*
- [x] 无基于行号的注释定位 - *测试将验证这一点*
- [x] 无正则表达式提取注释 - *测试将验证这一点*

**II. 单一语言源原则**:
- [x] Git 仓库只存储英文注释 - *测试数据将反映这一点*
- [x] 映射文件存储在 `.codei18n/` 且不提交 - *测试将验证文件生成位置*
- [x] pre-commit hook 实现注释语言转换 - *不在本次范围内，但兼容*

**III. 代码规范与测试**:
- [x] 遵循 Effective Go 规范
- [x] 单元测试覆盖率 ≥ 60%（核心模块 ≥ 80%） - *目标是提升覆盖率*
- [x] 使用 gofmt 和 golint/staticcheck
- [x] 核心功能有集成测试 - *本次主要任务*

**IV. 中文优先**:
- [x] 所有文档使用中文
- [x] 代码注释使用中文
- [x] 函数/类型命名使用英文（符合 Go 规范）

**V. CLI 优先**:
- [x] 核心功能通过 CLI 暴露
- [x] 支持 JSON 和人类可读输出 - *重点测试 JSON 输出*
- [x] IDE 插件仅作为渲染层 - *通过测试 JSON 协议保证*

## 项目结构

### 文档(此功能)

```
specs/004-integration-test-coverage/
├── plan.md              # 此文件
├── research.md          # 阶段 0 输出
├── data-model.md        # 阶段 1 输出
├── quickstart.md        # 阶段 1 输出
├── contracts/           # 阶段 1 输出
│   └── cli-output.schema.json
└── tasks.md             # 阶段 2 输出
```

### 源代码(仓库根目录)

```
tests/
├── integration_test.go          # 现有集成测试（将增强）
├── ide_protocol_test.go         # 新增：IDE 协议与 JSON 模式测试
├── cli_flags_test.go            # 新增：CLI 参数与配置优先级测试
├── complex_parsing_test.go      # 新增：复杂 Go 注释场景测试
├── rust_integration_test.go     # 新增：Rust 语言支持测试
└── fixtures/                    # 新增：测试用例数据
    ├── simple.go
    ├── complex.go               # 包含复杂注释的 Go 文件
    ├── lib.rs                   # 包含 Rust 特有注释的文件
    └── config.json
```

**结构决策**: 保持测试与源码分离，使用 `tests/` 目录存放集成测试，避免循环依赖，并模拟外部调用者视角。

## 复杂度跟踪

*仅在章程检查有必须证明的违规时填写*

| 违规 | 为什么需要 | 拒绝更简单替代方案的原因 |
|-----------|------------|-------------------------------------|
| (无) | | |
