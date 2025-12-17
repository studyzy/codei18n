# 快速开始: Integration Test Coverage

**功能分支**: `004-integration-test-coverage`

## 1. 运行集成测试

本项目使用 Go 标准测试工具链。

### 1.1 运行所有测试
包括单元测试和集成测试。

```bash
make test-integration
# 或者直接使用 go test
go test -v ./tests/...
```

### 1.2 仅运行新的 IDE 集成测试
(注意: 这些测试将在本功能分支中添加)

```bash
go test -v -run TestIDEPluginIntegration ./tests/...
```

### 1.3 运行带覆盖率报告的测试

```bash
make coverage
# 查看 HTML 报告
make coverage-html
```

## 2. 编写新测试

在 `tests/` 目录下创建新的 `*_test.go` 文件。

### 2.1 模板

```go
func TestNewFeature(t *testing.T) {
    // 1. 设置环境
    bin := buildBinary(t)
    tempDir := t.TempDir()

    // 2. 准备数据
    createFile(t, tempDir, "main.go", "...")

    // 3. 执行 CLI
    cmd := exec.Command(bin, "scan", "--file", "main.go", "--format", "json")
    cmd.Dir = tempDir
    output, err := cmd.CombinedOutput()

    // 4. 验证
    require.NoError(t, err)
    var result ScanResult
    err = json.Unmarshal(output, &result)
    require.NoError(t, err)
    assert.NotEmpty(t, result.Comments)
}
```

## 3. 常见问题

- **错误 `exec: "codei18n": executable file not found`**: 确保 `buildBinary` 辅助函数被正确调用且构建成功。
- **JSON 解析失败**: 检查 CLI 是否输出了非 JSON 的日志信息（如 `[INFO] Processing...`）。使用 `assert.NotContains(t, string(output), "[INFO]")` 进行调试。
