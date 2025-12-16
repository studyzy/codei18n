# 贡献指南

感谢您对 CodeI18n 项目的关注！我们欢迎所有形式的贡献。

## 行为准则

参与本项目即表示您同意遵守我们的行为准则。请友善、尊重他人。

## 如何贡献

### 报告问题

如果您发现了 bug 或有功能建议：

1. 先在 [Issues](https://github.com/studyzy/codei18n/issues) 中搜索是否已有相同问题
2. 如果没有，创建新的 Issue
3. 使用清晰的标题和详细的描述
4. 如果是 bug，请提供：
   - 操作系统和版本
   - Go 版本
   - 重现步骤
   - 预期行为和实际行为
   - 相关日志或截图

### 提交代码

#### 准备工作

1. Fork 项目到您的 GitHub 账户
2. Clone 您的 fork：
   ```bash
   git clone https://github.com/YOUR_USERNAME/codei18n.git
   cd codei18n
   ```
3. 添加上游仓库：
   ```bash
   git remote add upstream https://github.com/studyzy/codei18n.git
   ```
4. 安装依赖：
   ```bash
   make deps
   ```

#### 开发流程

1. 创建新分支：
   ```bash
   git checkout -b feature/your-feature-name
   ```
   
2. 进行开发，确保：
   - 遵循项目的代码规范（见下文）
   - 添加必要的测试
   - 更新相关文档

3. 提交前检查：
   ```bash
   # 运行所有检查
   make pre-commit
   
   # 或者运行完整 CI
   make ci
   ```

4. 提交代码：
   ```bash
   git add .
   git commit -m "feat: 添加新功能的简短描述"
   ```
   
   提交信息格式：
   - `feat: 新功能`
   - `fix: 修复 bug`
   - `docs: 文档更新`
   - `style: 代码格式化`
   - `refactor: 重构`
   - `test: 测试相关`
   - `chore: 构建/工具相关`

5. 推送到您的 fork：
   ```bash
   git push origin feature/your-feature-name
   ```

6. 在 GitHub 上创建 Pull Request

#### Pull Request 指南

- PR 标题要清晰简洁
- 在 PR 描述中：
  - 说明改动的目的和内容
  - 关联相关的 Issue（如 `Closes #123`）
  - 列出测试步骤
- 确保所有 CI 检查通过
- 保持 PR 专注于单一目的
- 及时回应评审意见

## 代码规范

### Go 代码规范

本项目严格遵循以下规范：

1. **Effective Go**：https://go.dev/doc/effective_go
2. **Go Code Review Comments**：https://go.dev/wiki/CodeReviewComments

### 强制要求

根据项目章程（`.specify/memory/constitution.md`）：

1. **代码格式化**
   ```bash
   # 必须通过 gofmt
   make fmt
   ```

2. **代码检查**
   ```bash
   # 必须通过 go vet
   make vet
   
   # 必须通过 staticcheck 或 golint
   make lint
   ```

3. **测试覆盖率**
   - 总体覆盖率 ≥ 60%
   - 核心模块（`core/comment`, `core/mapping`, `core/translate`）≥ 80%
   
   ```bash
   # 检查覆盖率
   make coverage-check
   ```

4. **中文优先**
   - 所有文档使用中文
   - 代码注释使用中文
   - 函数/类型命名使用英文（符合 Go 规范）

5. **AST 优先**
   - 所有注释处理必须基于 AST
   - 禁止使用正则表达式或基于行号的方式

### 项目结构

```
CodeI18n/
├── cmd/codei18n/        # CLI 入口
├── core/                # 核心业务逻辑
│   ├── comment/         # 注释模型
│   ├── mapping/         # 映射管理
│   └── translate/       # 翻译引擎
├── adapters/            # 语言适配器
│   └── go/             # Go 适配器
├── internal/            # 内部代码
└── tests/              # 测试
```

## 开发环境

### 必需工具

- Go 1.25.5+
- Make
- Git

### 推荐工具

- staticcheck：`go install honnef.co/go/tools/cmd/staticcheck@latest`
- govulncheck：`go install golang.org/x/vuln/cmd/govulncheck@latest`

### 常用命令

```bash
# 开发模式快速检查
make dev

# 运行所有测试
make test

# 查看覆盖率
make coverage-html

# 提交前完整检查
make pre-commit

# CI 完整流程
make ci
```

## 发布流程

仅限维护者：

1. 更新 CHANGELOG.md
2. 创建版本 tag：
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```
3. GitHub Actions 将自动构建和发布

## 获取帮助

- 阅读 [README.md](README.md)
- 查看 [快速开始指南](specs/001-codei18n-core-mvp/quickstart.md)
- 在 [Discussions](https://github.com/studyzy/codei18n/discussions) 提问
- 加入开发者社区（待建立）

## 许可证

提交代码即表示您同意在本项目的许可证（见 [LICENSE](LICENSE)）下贡献您的代码。

---

再次感谢您的贡献！🎉
