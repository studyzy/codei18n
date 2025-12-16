# 更新日志

本文档记录 CodeI18n 项目的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

### 新增
- 添加 GitHub Actions CI/CD 工作流
  - 代码质量检查（gofmt, go vet, staticcheck）
  - 多平台构建测试（Linux, macOS, Windows）
  - 单元测试和覆盖率报告
  - 集成测试
  - 安全扫描（gosec, govulncheck）
  - 依赖审计
- 添加自动发布工作流
  - 多平台二进制文件构建
  - Docker 镜像构建和发布
  - 自动生成 Release Notes
- 添加 Makefile 用于常用操作
- 添加 Dockerfile 支持容器化部署
- 添加依赖自动更新工作流

### 改进
- 更新 .gitignore 添加更多忽略模式
- 更新 README.md 添加开发工作流说明
- 优化构建流程

## [0.1.0] - TBD

### 新增
- Go 语言 AST 适配器
- 注释扫描和提取功能
- 注释 ID 生成逻辑
- 多语言映射管理
- 基于 Cobra 的 CLI 框架
- 基础配置管理

### 核心功能
- `scan` 命令：扫描源码并提取注释
- `init` 命令：初始化项目配置
- `map` 命令：管理多语言映射
- `translate` 命令：自动翻译注释
- `convert` 命令：转换注释语言
- `hook` 命令：管理 Git Hooks

[未发布]: https://github.com/studyzy/codei18n/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/studyzy/codei18n/releases/tag/v0.1.0
