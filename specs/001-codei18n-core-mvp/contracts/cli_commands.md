# CodeI18n CLI 命令规范

本文档定义了 CodeI18n 命令行工具的接口规范。CLI 使用 `cobra` 框架实现，遵循 Unix 哲学。

## 全局参数

所有命令均支持以下参数：

- `--config string`: 指定配置文件路径 (默认 `$HOME/.codei18n.json` 或项目根目录 `.codei18n/config.json`)
- `--verbose`: 启用详细日志输出
- `--help`: 显示帮助信息

## 1. Init 命令

初始化项目配置。

```bash
codei18n init [flags]
```

**Flags**:
- `--source-lang string`: 源码语言 (默认 "en")
- `--target-lang string`: 本地目标语言 (默认 "zh-CN")
- `--provider string`: 翻译提供商 ("google" 或 "openai", 默认 "google")

**行为**:
1. 检查当前目录是否已存在配置。
2. 创建 `.codei18n/` 目录。
3. 生成 `.codei18n/config.json` 文件。
4. 输出初始化成功信息。

## 2. Scan 命令

扫描源码并提取注释。

```bash
codei18n scan [flags]
```

**Flags**:
- `-f, --file string`: 指定扫描的单个文件路径
- `-d, --dir string`: 指定扫描的目录路径 (默认当前目录)
- `--format string`: 输出格式 ("json", "table", 默认 "table")
- `-o, --output string`: 将输出写入指定文件 (默认 stdout)
- `--stdin`: 从 stdin 读取文件内容（必须同时指定 `--file` 以确定语言和 ID 生成上下文）
- `--with-translations`: 在 JSON 输出中包含翻译文本（如果可用）

**行为**:
1. 如果指定 `--stdin`，从标准输入读取内容；否则遍历指定文件或目录。
2. 调用对应语言的适配器解析 AST。
3. 提取注释并生成 ID。
4. 如果指定 `--with-translations`，查询映射库并填充翻译。
5. 按指定格式输出结果。**注意：**当格式为 `json` 时，必须确保 stdout 只包含 JSON 数据，所有日志应写入 stderr。

**JSON 输出 Schema**:
```json
{
  "file": "path/to/main.go",
  "comments": [
    {
      "id": "e3b0c442...",
      "range": {
        "startLine": 10,
        "startCol": 1,
        "endLine": 12,
        "endCol": 3
      },
      "sourceText": "Calculate calculates sum",
      "translation": "Calculate 计算总和" // 仅在 --with-translations 时存在
    }
  ]
}
```

## 3. Map 命令

管理多语言映射。

```bash
codei18n map [subcommand] [flags]
```

### 3.1 Map Create/Update

```bash
codei18n map update [flags]
```

**Flags**:
- `--scan-dir string`: 扫描目录以更新映射 (默认当前目录)

**行为**:
1. 扫描项目中的所有注释。
2. 读取现有的 `mappings.json`。
3. 对比差异，添加新注释 ID，标记已删除的 ID。
4. 保存更新后的映射文件。

### 3.2 Map Get

```bash
codei18n map get <comment-id> --lang <lang-code>
```

**行为**:
输出指定 ID 和语言的翻译文本。

## 4. Translate 命令

调用翻译服务自动翻译缺失的条目。

```bash
codei18n translate [flags]
```

**Flags**:
- `--provider string`: 覆盖配置文件中的提供商
- `--model string`: 指定模型 (如 "gpt-3.5-turbo")
- `--concurrency int`: 并发请求数 (默认 5)

**行为**:
1. 加载映射文件。
2. 识别所有目标语言为空的条目。
3. 批量调用翻译 API。
4. 更新并保存映射文件。
5. 显示进度条和统计信息。

## 5. Convert 命令

在源码中执行注释语言转换（原地修改）。

```bash
codei18n convert [flags]
```

**Flags**:
- `-f, --file string`: 指定文件
- `-d, --dir string`: 指定目录
- `--to string`: 目标语言 ("en" 或 "zh-CN")
- `--dry-run`: 仅显示将要修改的内容，不实际写入

**行为**:
1. 扫描文件并匹配 AST 中的注释。
2. 根据 ID 在映射文件中查找目标语言文本。
3. 如果找到映射，替换源码中的注释内容。
4. 保持代码格式（利用 `go/format` 或 AST 节点位置）。

## 6. Hook 命令

管理 Git Hooks。

```bash
codei18n hook install
codei18n hook uninstall
```

**行为**:
- `install`: 在 `.git/hooks/pre-commit` 中写入调用脚本。
- `uninstall`: 移除对应的 hook 脚本。

## 错误处理

- 成功执行返回 Exit Code 0。
- 任何错误返回非零 Exit Code，并将错误信息写入 stderr。
