# cli Specification

## Purpose
TBD - created by archiving change enhance-init-command. Update Purpose after archive.
## 需求
### 需求：Init 命令增强

CLI **必须**提供增强的初始化命令，支持智能配置继承、自动映射构建和 Git 环境集成。

#### 场景：无全局配置初始化
当用户首次运行 `codei18n init` 且没有全局配置文件时：
1. **Given** 用户在没有任何配置的项目目录中
2. **When** 运行 `codei18n init`
3. **Then** **必须**在 `.codei18n/config.json` 创建默认配置
4. **And** **必须**提示用户可以配置全局 API Key 以便后续项目复用

#### 场景：继承全局配置
当用户已配置全局文件（如 `~/.codei18n/config.json`）时：
1. **Given** 存在包含 `provider` 和 `api_key` 的全局配置
2. **When** 运行 `codei18n init`
3. **Then** 项目级 `.codei18n/config.json` **必须**继承全局配置的非敏感字段（如 `source_lang`, `target_lang`）
4. **And** 项目级配置**必须禁止**包含 `api_key`（除非用户显式要求或通过 flag 覆盖），以确保安全

#### 场景：初始化时扫描代码
用户希望初始化后立即能看到翻译效果：
1. **Given** 一个包含源代码的现有项目
2. **When** 运行 `codei18n init`
3. **Then** 命令**必须**自动执行等同于 `codei18n map update` 的逻辑
4. **And** **必须**在 `.codei18n/` 目录下生成初始的 `mappings.json`
5. **And** **必须**输出扫描到的注释数量

#### 场景：带翻译的初始化
用户希望一步到位生成中文注释视图：
1. **Given** 项目未初始化
2. **When** 运行 `codei18n init --with-translate`
3. **Then** **必须**完成配置生成和映射构建
4. **And** **必须**自动调用翻译引擎（如 OpenAI）对提取的注释进行翻译
5. **And** **必须**更新 `mappings.json` 填入翻译结果

#### 场景：Git 仓库自动配置
当用户在 Git 仓库中初始化时：
1. **Given** 当前目录是一个 Git 仓库（存在 `.git` 目录）
2. **When** 运行 `codei18n init`（包含或不包含 `--with-translate`）
3. **Then** **必须**检查 `.gitignore` 文件
4. **And** 如果 `.gitignore` 中不包含 `.codei18n/`，**必须**追加该条目（如果文件不存在则创建）
5. **And** **必须**自动执行 `codei18n hook install` 以安装 pre-commit hook
6. **And** 如果当前目录不是 Git 仓库，则**必须**跳过上述步骤

### 需求：接口定义更新

CLI 接口和标志定义**必须**更新以支持新的初始化流程。

#### Flags 更新

```bash
codei18n init [flags]
```

新增/更新 Flags:
- `--with-translate`: 初始化后立即执行翻译（默认 false）
- `--provider string`: 指定翻译提供商（覆盖全局配置）
- `--force`: 如果配置已存在，强制覆盖

#### 执行流程

1. **配置加载阶段**
   - 尝试读取全局配置
   - 合并 CLI 参数
   - 生成项目级 Config 对象（过滤敏感字段）
   - 写入 `.codei18n/config.json`

2. **Git 环境配置阶段**
   - 检查是否为 Git 仓库
   - 更新/创建 `.gitignore` 添加 `.codei18n/`
   - 执行 `codei18n hook install`

3. **映射构建阶段**
   - 扫描当前目录（排除 `.codei18n` 和 `.git`）
   - 提取所有注释并生成 ID
   - 创建/更新 `mappings.json`

4. **翻译阶段（可选）**
   - 如果指定 `--with-translate`：
     - 初始化 Translator
     - 识别未翻译条目
     - 执行批量翻译
     - 保存映射

