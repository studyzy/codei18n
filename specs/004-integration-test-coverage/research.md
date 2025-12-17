# 研究: Integration Test Coverage

**功能**: Integration Test Coverage
**状态**: 完成
**日期**: 2025-12-17

## 1. 技术上下文分析

### 1.1 现有测试基础设施
项目目前使用标准的 Go `testing` 包配合 `github.com/stretchr/testify` 进行测试。
- **单元测试**: 广泛分布在各包中。
- **集成测试**: 存在于 `tests/integration_test.go`，采用 "构建二进制 -> exec.Command 执行 -> 验证输出" 的模式。
- **Mock**: 已有 `adapters/translator/mock_llm_test.go` 和 CLI层面的 `--provider mock` 支持。

### 1.2 关键依赖
- **CLI 框架**: Cobra
- **配置管理**: Viper
- **断言库**: Testify (assert/require)
- **外部进程执行**: `os/exec` (标准库)

## 2. 决策记录

### 2.1 测试框架选择
**决策**: 继续使用 Go 标准库 `testing` + `testify` + `os/exec` 模式，不引入新的测试框架（如 `testscript` 或 `ginkgo`）。

**理由**:
1. **一致性**: 保持与现有 `tests/integration_test.go` 的模式一致，降低维护成本。
2. **灵活性**: `os/exec` 提供了对 CLI 进程的最真实模拟（包括 stdin/stdout 管道、退出代码、信号处理）。
3. **无额外依赖**: 避免为了测试引入重量级框架。

**备选方案**:
- `rogpeppe/go-internal/testscript`: Go 团队使用的脚本化测试工具。虽然强大，但学习曲线较高，且引入了新的 DSL，不如纯 Go 代码直观。

### 2.2 IDE 插件模拟策略
**决策**: 创建专门的测试套件 `TestIDEPluginIntegration`，模拟 IDE 插件的调用行为。

**具体策略**:
1. **JSON 严格校验**: IDE 插件极其依赖 JSON 输出格式。测试必须反序列化 CLI 的 stdout 到结构体，并验证字段的完整性和类型。
2. **Stdin 输入支持**: IDE 插件常通过 stdin 传递脏缓冲区（unsaved buffer）的内容。测试需验证 `codei18n scan --stdin` 场景。
3. **错误处理**: 验证 CLI 在错误时（如解析失败）是否向 stderr 输出结构化或可解析的错误信息，且退出代码正确。

### 2.3 测试数据管理
**决策**: 使用 `embed` 包或行内字符串定义 Test Fixtures，在 `t.TempDir()` 中动态生成文件。

**理由**:
- **隔离性**: 每个测试使用独立临时目录，互不干扰。
- **便携性**: 测试代码包含数据，无需依赖外部文件路径。

## 3. 风险与缓解

- **构建时间**: 每次测试都 `go build` 会很慢。
  - **缓解**: 使用 `TestMain` 进行一次性构建，所有测试复用同一个二进制文件。
- **平台差异**: Windows 下的文件路径和换行符可能不同。
  - **缓解**: 使用 `filepath.Join` 和通用断言逻辑，避免硬编码路径分隔符。

## 4. 结论
我们将扩展现有的 `tests/` 目录，增加覆盖更多 CLI 参数和边缘情况的集成测试。重点在于验证 "IDE 协议"（即 CLI 参数输入和 JSON 输出）的稳定性。
