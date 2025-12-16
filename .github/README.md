# GitHub Actions 工作流说明

本目录包含 CodeI18n 项目的 GitHub Actions 工作流配置。

## 工作流概览

### 1. CI 工作流 (`ci.yml`)

**触发条件**：
- Push 到 `main`、`master`、`develop` 分支
- Pull Request 到上述分支

**包含的检查**：

#### 代码质量检查 (lint)
- Go mod 验证
- gofmt 格式检查
- go vet 静态分析
- staticcheck 代码检查

#### 多平台构建 (build)
- 测试平台：Ubuntu、macOS、Windows
- Go 版本：1.25.5
- 验证构建产物

#### 单元测试 (test)
- 运行所有测试（带竞态检测）
- 生成覆盖率报告
- 验证覆盖率要求（≥60%）
- 上传到 Codecov

#### 集成测试 (integration-test)
- CLI 命令测试
- 端到端功能验证

#### 安全扫描 (security)
- gosec 安全扫描
- govulncheck 漏洞检查

#### 依赖审计 (dependencies)
- 检查过时的依赖
- 验证依赖完整性

#### 许可证检查 (license)
- 验证 LICENSE 文件存在

### 2. Release 工作流 (`release.yml`)

**触发条件**：
- 推送符合 `v*.*.*` 格式的 tag

**执行任务**：

#### 构建多平台二进制文件
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

#### 生成发布资源
- 压缩包（.tar.gz 和 .zip）
- SHA256 校验和
- Release Notes

#### Docker 镜像（可选）
- 多平台镜像构建
- 推送到 Docker Hub

**示例**：

```bash
# 创建并推送 release tag
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# GitHub Actions 将自动：
# 1. 运行所有测试
# 2. 构建多平台二进制文件
# 3. 创建 GitHub Release
# 4. 构建并推送 Docker 镜像
```

### 3. 依赖更新工作流 (`dependency-update.yml`)

**触发条件**：
- 每周一 UTC 00:00 自动运行
- 手动触发

**执行任务**：
- 更新所有 Go 依赖到最新版本
- 运行 `go mod tidy`
- 运行测试验证
- 自动创建 Pull Request

## 使用说明

### 本地验证 CI

在推送代码前，可以在本地运行与 CI 相同的检查：

```bash
# 运行所有 CI 检查
make ci

# 或者分步运行
make fmt          # 格式化检查
make vet          # 静态分析
make lint         # 代码检查
make test         # 单元测试
make coverage-check  # 覆盖率检查
```

### 查看 CI 状态

- 在 PR 页面查看所有检查的状态
- 点击 "Details" 查看详细日志
- 所有检查必须通过才能合并

### 发布新版本

1. 确保所有 CI 检查通过
2. 更新 CHANGELOG.md
3. 创建 tag：
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```
4. 等待 GitHub Actions 完成构建
5. 检查 [Releases](https://github.com/studyzy/codei18n/releases) 页面

### Docker 镜像发布

如果需要发布 Docker 镜像，需要配置以下 secrets：

1. 进入仓库 Settings → Secrets and variables → Actions
2. 添加以下 secrets：
   - `DOCKER_USERNAME`: Docker Hub 用户名
   - `DOCKER_PASSWORD`: Docker Hub 访问令牌

然后推送版本 tag 即可自动构建和推送镜像。

### Codecov 集成（可选）

如果需要 Codecov 集成：

1. 访问 https://codecov.io/
2. 添加您的仓库
3. 获取 CODECOV_TOKEN
4. 在仓库 Settings → Secrets 中添加 `CODECOV_TOKEN`

## Issue 和 PR 模板

### Issue 模板

位于 `ISSUE_TEMPLATE/` 目录：

- `bug_report.md`: Bug 报告模板
- `feature_request.md`: 功能请求模板

创建 Issue 时会自动显示模板选项。

### PR 模板

位于 `pull_request_template.md`，创建 PR 时自动加载。

包含：
- 变更说明
- 类型选择
- 检查清单
- 测试步骤

## 故障排查

### CI 失败常见原因

1. **格式化检查失败**
   ```bash
   make fmt
   git add .
   git commit --amend
   ```

2. **测试失败**
   ```bash
   make test-verbose  # 查看详细错误
   ```

3. **覆盖率不足**
   ```bash
   make coverage-html  # 查看覆盖率详情
   ```

4. **Lint 检查失败**
   ```bash
   make lint  # 查看具体问题
   ```

### Docker 构建失败

```bash
# 本地测试 Docker 构建
make docker-build

# 查看构建日志
docker build --no-cache -t codei18n:test .
```

## 维护

### 更新工作流

修改 `.github/workflows/*.yml` 文件后：

1. 在本地测试相关命令
2. 提交并推送到 feature 分支
3. 创建 PR 并观察 CI 结果
4. 合并后观察工作流执行

### 监控

- 定期检查 [Actions](https://github.com/studyzy/codei18n/actions) 页面
- 关注失败的工作流
- 及时修复问题

---

更多信息请参考：
- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [项目贡献指南](../CONTRIBUTING.md)
