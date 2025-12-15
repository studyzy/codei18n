# 技术调研: CodeI18n 技术栈

## CLI 框架 (CLI Framework)
**决策**: [spf13/cobra](https://github.com/spf13/cobra)
**理由**: Cobra 是构建 Go 命令行应用程序的事实上的行业标准。Kubernetes、Hugo 和 GitHub CLI 都使用它。它开箱即用地提供了对子命令（scan, map, convert 等）、标志（flags）和帮助文档生成的强大支持。
**替代方案**:
- `urfave/cli`: 一个强有力的竞争者，但 Cobra 拥有更大的生态系统，并且与 Viper（配置管理）的集成更好。
- `flag` (标准库): 对于具有子命令的复杂 CLI 来说过于原始。

## 配置管理 (Configuration)
**决策**: [spf13/viper](https://github.com/spf13/viper)
**理由**: Viper 是 Cobra 的标准伴侣。它可以无缝处理 JSON/YAML 文件、环境变量和命令行标志的读取。它支持默认值和嵌套配置，非常适合我们 `ide.vscode` 的结构配置需求。
**替代方案**:
- `kelseyhightower/envconfig`: 对于 12-factor 应用（仅使用环境变量）非常出色，但我们需要支持配置文件（JSON/YAML）来管理项目特定的设置。
- `knadh/koanf`: 一个更轻量、现代的替代品，但 Viper 的社区支持和 Cobra 集成使其成为 MVP 的更安全选择。

## 翻译客户端 (Translation Clients)

### LLM Client (OpenAI / DeepSeek)
**决策**: [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai)
**理由**:
- **兼容性**: DeepSeek 及国内主要大模型（如 Moonshot）均完全兼容 OpenAI API 规范。仅需配置 `BaseURL` 即可无缝切换，无需引入额外抽象层。
- **轻量级**: 相比于 `tmc/langchaingo`，该库专注且依赖少，适合作为 CLI 工具的核心依赖。
- **社区标准**: Go 生态中最成熟的 OpenAI 客户端，维护活跃。
**替代方案**:
- `tmc/langchaingo`: Go 语言的 LangChain 实现。虽然支持多模型抽象（如同时支持 Anthropic/Gemini 原生接口），但对于 MVP 阶段仅需文本翻译的场景过于厚重。如果未来需要复杂的 Agent 逻辑或引入非 OpenAI 兼容模型，再考虑迁移。
- `net/http`: 需要处理重试、速率限制和类型定义，样板代码过多。

### Google Translate
**决策**: [cloud.google.com/go/translate](https://pkg.go.dev/cloud.google.com/go/translate)
**理由**: Go 语言的官方 Google Cloud 客户端库。它提供了访问翻译 API（基础版和高级版）最可靠和最新的方式。
**替代方案**:
- REST API 调用: 与官方 SDK 相比，类型安全性较差，且需要手动处理身份验证。

## 日志与 CLI 交互 (Logging & CLI UX)
**决策**: [fatih/color](https://github.com/fatih/color) (颜色) & [briandowns/spinner](https://github.com/briandowns/spinner) (加载动画)
**理由**:
- `fatih/color`: Go 语言 ANSI 颜色输出的标准库。用于高亮成功/错误消息，简单有效。
- `briandowns/spinner`: 对于网络请求（翻译）或大文件扫描等长时间运行的操作，提供反馈至关重要。
- `rs/zerolog`: 用于内部结构化日志（调试日志），因为它是零分配且速度很快。
**替代方案**:
- `charmbracelet/bubbletea`: 非常适合全 TUI（终端用户界面）应用，但对于需要与 Git hooks 配合的简单 CLI 来说有点杀鸡用牛刀。

## 测试 (Testing)
**决策**: [stretchr/testify](https://github.com/stretchr/testify)
**理由**: 断言（`assert`, `require`）和模拟（`mock`）的行业标准。与标准 `testing` 包的检查相比，它大大减少了样板代码，这对于满足 60%+ 覆盖率要求至关重要。
**替代方案**:
- `matryer/is`: 极简主义，但 Testify 的套件和 Mock 生成工具对于测试复杂的逻辑（如 AST 解析和 API 集成）更为强大。
- 标准 `testing`: 对于广泛的单元测试来说过于冗长。

## JSON 处理 (JSON Handling)
**决策**: 标准库 `encoding/json`
**理由**: 标准库健壮、稳定，对于配置和映射文件来说已经足够。我们的 JSON 结构（映射文件）定义明确，且不是特别大（目前还不需要流式处理）。
**替代方案**:
- `json-iterator/go`: 更快，但会引入依赖。如果稍后性能成为瓶颈，可以进行替换。

## IDE 集成策略 (IDE Integration Strategy)
**决策**: CLI 执行 + Stdin/Stdout (基于单次 CLI 调用的 JSON-RPC 风格)
**理由**:
- **复杂度 vs 价值**: 实现完整的语言服务器 (LSP) 会显著增加复杂度（连接管理、状态同步、协议遵循）。对于 MVP，主要需求是“扫描文件 -> 获取注释”和“获取 ID 的翻译”。无状态的 CLI 构建和维护要简单得多。
- **性能**: Go 的启动速度极快。对于 MVP，在文件保存或去抖动（debounce）时生成 `scan` 进程是可以接受的。
- **灵活性**: VS Code 和 JetBrains 都有强大的 API 用于生成进程（Node 中的 `cp.spawn`，Java/Kotlin 中的 `GeneralCommandLine`）。

**新增需求**:

1.  **`scan` 命令支持 Stdin**:
    - **需求**: IDE 通常需要扫描“脏”缓冲区（未保存的更改）。
    - **解决方案**: `scan` 命令必须支持通过 Stdin 接收内容。
    - **标志**: `--stdin` (布尔值) 或自动检测。
    - **上下文**: 当从 Stdin 扫描时，仍然需要 `--file` 参数来确定语言解析器并生成稳定的 ID（ID 生成依赖于文件路径）。

2.  **严格的 JSON 输出 Schema**:
    - **需求**: IDE 需要精确的位置数据来渲染装饰器 (VS Code) 或 Inlay Hints (JetBrains)。
    - **Schema 示例**:
      ```json
      {
        "file": "path/to/file.go",
        "comments": [
          {
            "id": "a1b2c3d4",
            "range": { "start_line": 10, "start_col": 4, "end_line": 10, "end_col": 20 },
            "original": "Calculates sum",
            "translation": "计算总和" // 可选，如果已加载映射
          }
        ]
      }
      ```

3.  **Lookup 命令**:
    - **需求**: IDE 需要快速获取特定 ID 的翻译以显示在悬停提示或内嵌提示中。
    - **优化**: `codei18n map get <id>` 已经足够，但确保 `scan` 可以选择性地一次性返回翻译（例如 `--with-translations`）可以减少 N+1 次 CLI 调用。

4.  **错误处理**:
    - JSON 输出必须是纯净的。日志消息 (info/debug) 必须输出到 `stderr`。只有结果 JSON 输出到 `stdout`。

**未来考虑**:
- 如果在大文件或高频输入时延迟成为问题，迁移到“守护进程模式”（监听 stdin/socket 的长运行进程）或完整的 LSP。
