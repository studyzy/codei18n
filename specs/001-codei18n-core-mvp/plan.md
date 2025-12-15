# 实施计划: CodeI18n Core MVP

**分支**: `001-codei18n-core-mvp` | **日期**: 2025-12-15 | **规范**: [specs/001-codei18n-core-mvp/spec.md](specs/001-codei18n-core-mvp/spec.md)
**输入**: 来自 `specs/001-codei18n-core-mvp/spec.md` 的功能规范

**注意**: 此模板由 `/speckit.plan` 命令填充. 执行工作流程请参见 `.specify/templates/commands/plan.md`.

## 摘要

本项目旨在构建 CodeI18n 的核心 MVP，实现 Go 语言源代码注释的扫描、ID 生成、多语言映射管理以及基于 Google Translate 和 OpenAI/DeepSeek 的自动翻译功能。通过 CLI 工具和 pre-commit hook，实现"源码英文、本地母语"的无缝切换体验。技术选型采用 Go 语言，基于 Cobra 构建 CLI，使用 Viper 管理配置，利用标准库 `go/ast` 进行 AST 解析，确保注释与语义的稳定绑定。**特别地，设计针对 IDE 集成进行了优化，支持 Stdin 输入和严格的 JSON 输出。**

## 技术背景

<!--
  需要操作: 将此部分内容替换为项目的技术细节.
  此处的结构以咨询性质呈现, 用于指导迭代过程.
-->

**语言/版本**: Go 1.25+
**主要依赖**:
- CLI: `spf13/cobra` (命令框架), `spf13/viper` (配置管理)
- UX: `fatih/color` (彩色输出), `briandowns/spinner` (进度提示)
- AI/翻译: `sashabaranov/go-openai` (OpenAI/DeepSeek 客户端), `cloud.google.com/go/translate` (Google 翻译)
- 测试: `stretchr/testify` (断言与 Mock)
- AST: 标准库 `go/parser`, `go/ast` (无第三方依赖)
**存储**: 本地 JSON 文件 (`.codei18n/mappings.json`, `.codei18n/config.json`)
**测试**: 单元测试覆盖率 > 60%，使用 Testify 进行表驱动测试
**目标平台**: macOS, Linux, Windows
**项目类型**: CLI 工具 + 本地库
**性能目标**: 单文件扫描 < 100ms, 项目扫描 < 1s
**约束条件**: 必须使用 AST 解析，禁止正则表达式提取注释
**IDE 集成约束**:
- 支持 Stdin 扫描（用于 IDE 脏缓冲区）
- JSON 输出必须与日志（Stderr）严格分离
- 提供 `--with-translations` 选项以减少 CLI 调用次数

## 章程检查

*门控: 必须在阶段 0 研究前通过. 阶段 1 设计后重新检查. *

**I. AST 优先原则**:
- [x] 所有注释处理使用 AST（go/parser、go/ast、syn、tree-sitter）
- [x] 无基于行号的注释定位
- [x] 无正则表达式提取注释

**II. 单一语言源原则**:
- [x] Git 仓库只存储英文注释
- [x] 映射文件存储在 `.codei18n/` 且不提交
- [x] pre-commit hook 实现注释语言转换

**III. 代码规范与测试**:
- [x] 遵循 Effective Go 规范
- [x] 单元测试覆盖率 ≥ 60%（核心模块 ≥ 80%）
- [x] 使用 gofmt 和 golint/staticcheck
- [x] 核心功能有集成测试

**IV. 中文优先**:
- [x] 所有文档使用中文
- [x] 代码注释使用中文
- [x] 函数/类型命名使用英文（符合 Go 规范）

**V. CLI 优先**:
- [x] 核心功能通过 CLI 暴露
- [x] 支持 JSON 和人类可读输出
- [x] IDE 插件仅作为渲染层

## 项目结构

### 文档(此功能)

```
specs/001-codei18n-core-mvp/
├── plan.md              # 此文件
├── research.md          # 技术调研与选型
├── data-model.md        # 核心数据模型 (Comment, Mapping, Config)
├── quickstart.md        # 快速开始指南
├── contracts/           # 接口定义与 CLI 规范
│   ├── interfaces.go    # Go 接口定义
│   └── cli_commands.md  # CLI 命令规范
└── tasks.md             # 待执行任务列表
```

### 源代码(仓库根目录)

```
cmd/
└── codei18n/            # CLI 入口 (main.go)

core/                    # 核心业务逻辑
├── comment/             # 注释模型 (Model)
├── mapping/             # 映射文件管理 (Store)
└── translate/           # 翻译引擎 (Service)

adapters/                # 语言适配器 (Port implementation)
└── go/                  # Go 语言 AST 适配器

internal/                # 内部私有代码
├── cli/                 # Cobra 命令实现 (scan, map, etc.)
└── config/              # Viper 配置加载

pkg/                     # 公共工具库 (可选)
└── utils/               # 通用工具函数

tests/                   # 集成测试与测试数据
├── integration/         # CLI 端到端测试
└── testdata/            # 测试用例代码文件
```

**结构决策**: 采用标准的 Go 项目结构，结合 `CODEBUDDY.md` 中定义的 `core/` 和 `adapters/` 分层架构。`cmd` 存放入口，`internal` 存放应用特定逻辑（CLI 实现），`core` 定义核心领域模型和接口，`adapters` 实现具体的语言解析逻辑。

## 复杂度跟踪

*仅在章程检查有必须证明的违规时填写*

| 违规 | 为什么需要 | 拒绝更简单替代方案的原因 |
|-----------|------------|-------------------------------------|
| 无 | 本计划严格遵循章程 | N/A |
