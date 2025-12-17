# 测试指南

## 测试类型

本项目包含两种类型的测试：

### 1. 单元测试和集成测试（默认）

这些测试不需要外部依赖，可以随时运行：

```bash
# 运行所有测试（跳过外部 API 集成测试）
make test

# 或直接使用 go test
go test -short ./...
```

### 2. 外部 API 集成测试

某些测试需要真实的外部 API（如 DeepSeek、OpenAI 等）。这些测试默认被跳过，需要手动启用。

#### 运行 API 集成测试

**前提条件**：
- 设置有效的 `OPENAI_API_KEY` 环境变量
- 设置正确的 `OPENAI_BASE_URL`（可选，默认为 DeepSeek API）

```bash
# 运行所有测试（包括 API 集成测试）
make test-integration

# 或直接使用 go test（不带 -short 标志）
go test ./...
```

**示例**：

```bash
# 使用 DeepSeek API
export OPENAI_API_KEY="your-deepseek-api-key"
export OPENAI_BASE_URL="https://api.deepseek.com"
export OPENAI_MODEL="deepseek-chat"
make test-integration

# 使用 OpenAI API
export OPENAI_API_KEY="your-openai-api-key"
export OPENAI_BASE_URL="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-4"
make test-integration
```

## 测试覆盖率

```bash
# 生成覆盖率报告（跳过集成测试）
make coverage

# 查看 HTML 格式的覆盖率报告
make coverage-html

# 检查覆盖率是否达标（总体 ≥60%，核心模块 ≥80%）
make coverage-check
```

## CI/CD 测试

在 CI/CD 环境中，应该只运行不依赖外部 API 的测试：

```bash
# CI/CD 推荐
make test
make coverage-check
```

## 开发工作流

```bash
# 开发过程中的快速检查
make dev

# 提交前检查
make pre-commit
```

## 测试文件位置

- `adapters/rust/*_test.go` - Rust 适配器测试
- `adapters/translator/*_test.go` - 翻译器测试（包含 API 集成测试）
- `internal/log/*_test.go` - 日志模块测试
- `tests/*_test.go` - 端到端集成测试

## 注意事项

1. **API 集成测试**：`TestDeepSeekIntegration` 是一个手动测试，用于验证外部 API 连接。它在以下情况会被跳过：
   - 使用 `-short` 标志运行测试时
   - 未设置 `OPENAI_API_KEY` 环境变量时

2. **测试隔离**：所有单元测试都使用 mock 对象，不依赖外部服务。

3. **测试覆盖率要求**：
   - 总体覆盖率：≥ 60%
   - 核心模块（`core/comment`、`core/mapping`、`core/translate`）：≥ 80%
