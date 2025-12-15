# CodeI18n 快速开始指南

欢迎使用 CodeI18n！本指南将帮助你快速在 Go 项目中设置和使用 CodeI18n，实现代码注释的自动国际化。

## 前置条件

- Go 1.25+
- Git 项目

## 1. 安装

```bash
go install github.com/studyzy/codei18n/cmd/codei18n@latest
```

验证安装：
```bash
codei18n --version
```

## 2. 初始化项目

在你的 Go 项目根目录下运行：

```bash
codei18n init --source-lang en --target-lang zh-CN
```

这将创建 `.codei18n/config.json` 配置文件。

## 3. 扫描注释与创建映射

首次使用，需要扫描项目中的现有注释并创建映射文件：

```bash
# 扫描当前目录并初始化映射文件
codei18n map update
```

此时 `.codei18n/mappings.json` 文件已生成，其中包含所有扫描到的注释 ID，但中文翻译部分可能为空（如果你尚未翻译）。

## 4. 自动翻译

配置好翻译 API Key（推荐使用环境变量）：

### 使用 OpenAI

```bash
export OPENAI_API_KEY="sk-..."
```

### 使用 DeepSeek (或其他 OpenAI 兼容接口)

CodeI18n 支持所有兼容 OpenAI 接口的模型（如 DeepSeek, Moonshot 等）。
你可以在环境变量中设置 Base URL 和 API Key：

```bash
export OPENAI_API_KEY="sk-..."
export OPENAI_BASE_URL="https://api.deepseek.com/v1"
```

或者在 `.codei18n/config.json` 中配置：
```json
{
  "translationProvider": "openai",
  "translationConfig": {
    "baseUrl": "https://api.deepseek.com/v1",
    "model": "deepseek-chat"
  }
}
```

### 使用 Google Translate

```bash
export GOOGLE_APPLICATION_CREDENTIALS="path/to/key.json"
```

运行自动翻译命令，填充映射文件中的中文翻译：

```bash
codei18n translate --provider openai
```

系统会自动将所有英文注释翻译为中文并保存到 `mappings.json`。

## 5. 本地查看中文注释

想在本地开发时查看中文注释？

```bash
codei18n convert --dir . --to zh-CN
```

现在打开你的源码文件，所有注释都变成了中文！

## 6. 还原为英文（准备提交）

在提交代码到 Git 之前，或者想要还原为英文时：

```bash
codei18n convert --dir . --to en
```

## 7. 配置 Git Hook (推荐)

为了防止意外将中文注释提交到仓库，安装 pre-commit hook：

```bash
codei18n hook install
```

现在，每次执行 `git commit` 时，工具会自动检查并将暂存区（staged）文件中的注释转换为英文。

## 下一步

- 查看 `specs/001-codei18n-core-mvp/spec.md` 了解完整功能规范。
- 探索 `codei18n help` 查看更多命令选项。
