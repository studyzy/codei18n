# CodeI18n Core MVP 实施任务清单

本项目遵循 [CodeBuddy Code 规范](../../CODEBUDDY.md)。所有代码注释和文档必须使用中文。
所有输出到 Stdout 的数据（在 JSON 模式下）必须是纯净的 JSON，日志必须输出到 Stderr。

## 阶段 1: 设置与基础设施 (Setup & Infrastructure)

- [ ] **S001**: 初始化 Go 模块和基本目录结构 <!-- id: 0 -->
    - 执行 `go mod init github.com/studyzy/codei18n`
    - 创建目录: `cmd/codei18n`, `core`, `adapters/go`, `internal/utils`, `pkg`
    - 验证: `go mod tidy` 无报错
- [ ] **S002**: 搭建 Cobra 和 Viper CLI 骨架 <!-- id: 1 -->
    - 创建 `cmd/codei18n/root.go`: 定义根命令
    - 集成 `viper` 读取配置文件 (`.codei18n/config.json`)
    - 验证: `go run ./cmd/codei18n --help` 显示帮助信息
- [ ] **S003**: 配置日志系统 (Stdout/Stderr 分离) <!-- id: 2 -->
    - 实现 `internal/log/logger.go`
    - 确保常规日志输出到 **Stderr** (使用 `fatih/color` 或 `zap`)
    - 确保 **Stdout** 仅用于 CLI 命令的结果输出 (如 JSON)
    - 验证: 编写测试脚本验证 `fmt.Println` 和 `log.Println` 的输出流向不同
- [ ] **S004**: 定义全局配置结构与加载逻辑 <!-- id: 3 -->
    - 创建 `core/config/config.go`
    - 定义 `Config` 结构体 (包含 `SourceLang`, `TargetLang`, `IDE` 设置)
    - 实现 `LoadConfig` 和 `SaveConfig`
    - 验证: 能够读取并解析示例配置文件

## 阶段 2: 基础与核心接口 (Foundations & Core Interfaces)

- [ ] **F001**: 定义核心领域模型 (Comment & Mapping) <!-- id: 4 -->
    - 创建 `core/domain/model.go`
    - 定义 `Comment` 结构体 (ID, File, Language, Symbol, Range, SourceText)
    - 定义 `LocalizedComment` 结构体 (CommentID, Lang, Text)
    - 定义 `Mapping` 结构体 (KV 存储结构)
- [ ] **F002**: 定义核心接口 (Adapter & Translator) <!-- id: 5 -->
    - 创建 `core/interfaces.go`
    - 定义 `LanguageAdapter` 接口 (`Parse(file string, src []byte)`)
    - 定义 `Translator` 接口 (`Translate(text, from, to)`)
- [ ] **F003**: 实现映射存储逻辑 (Mapping Store) <!-- id: 6 -->
    - 创建 `core/mapping/store.go`
    - 实现 JSON 文件的读写逻辑 (支持并发读写锁)
    - 实现 `Get(id)`, `Set(id, text)`, `Save()` 方法
    - 验证: 单元测试覆盖读写操作

## 阶段 3: 扫描与 IDE 支持 (Scanning & IDE Support)

- [ ] **P101**: 实现 Go 语言 AST 适配器 <!-- id: 7 -->
    - 创建 `adapters/go/parser.go`
    - 使用 `go/parser` 和 `go/ast` 解析代码
    - 提取注释 (`//` 和 `/* */`) 及其关联的 AST 节点信息 (函数名/结构体名)
    - **关键**: 必须支持传入 `src []byte` 以处理 IDE 的脏缓冲区 (Dirty Buffer)
    - 验证: 单元测试覆盖不同类型的 Go 注释提取
- [ ] **P102**: 实现注释 ID 生成逻辑 <!-- id: 8 -->
    - 创建 `core/comment/id.go`
    - 实现 SHA1 算法: `SHA1(file_path + language + parent_symbol + normalized_text)`
    - 验证: 修改代码行号不影响 ID，修改函数名或注释内容会改变 ID
- [ ] **P103**: 实现 `scan` 命令基础逻辑 <!-- id: 9 -->
    - 创建 `cmd/codei18n/scan.go`
    - 遍历指定文件或目录，调用 Adapter 解析
    - 输出提取到的 `Comment` 列表
- [ ] **P104**: 实现 `scan` 命令的 Stdin 支持与 JSON 输出 <!-- id: 10 -->
    - 更新 `scan` 命令支持从 Stdin 读取内容 (用于 IDE 实时预览)
    - 实现 `--format json` 参数
    - **关键**: 确保 JSON 输出是标准合法的，不包含任何日志杂音
    - 验证: `cat file.go | ./codei18n scan --format json --stdin` 输出合法 JSON

## 阶段 4: 映射管理 (Mapping Management)

- [ ] **P201**: 实现 `map update` 命令 <!-- id: 11 -->
    - 创建 `cmd/codei18n/map.go`
    - 逻辑: 扫描代码 -> 提取新注释 -> 更新 Mapping Store -> 保存文件
    - 支持 `--dry-run`
- [ ] **P202**: 实现 `scan --with-translations` <!-- id: 12 -->
    - 更新 `scan` 命令
    - 在扫描时从 Mapping Store 查找对应的翻译
    - 将翻译结果填充到输出中 (供 IDE 插件渲染)
    - 验证: 扫描结果包含 `localizedText` 字段

## 阶段 5: 翻译引擎 (Translation Engine)

- [ ] **P301**: 实现 Mock 翻译器与接口集成 <!-- id: 13 -->
    - 创建 `adapters/translator/mock.go` (用于测试，直接返回 "Translated: " + text)
    - 在 Config 中支持配置翻译器类型
- [ ] **P302**: 实现 OpenAI/DeepSeek 翻译适配器 <!-- id: 14 -->
    - 创建 `adapters/translator/openai.go`
    - 实现 API 调用逻辑 (支持 Batch 请求以节省 Token)
    - 添加重试机制
- [ ] **P303**: 实现 `translate` 命令 <!-- id: 15 -->
    - 创建 `cmd/codei18n/translate.go`
    - 逻辑: 读取 Mapping 中未翻译的条目 -> 调用 Translator -> 更新 Mapping -> 保存
    - 支持 `--concurrency` 并发控制

## 阶段 6: 转换与 Hook (Conversion & Hooks)

- [ ] **P401**: 实现 `convert` 命令 (Apply/Restore) <!-- id: 16 -->
    - 创建 `cmd/codei18n/convert.go`
    - 支持 `apply`: 将源码注释替换为目标语言 (不推荐，除非特定需求)
    - 支持 `restore`: 将源码注释恢复为英文 (从 Mapping 中找回)
    - 验证: 双向转换测试，确保代码无损
- [ ] **P402**: 实现 Git Pre-commit Hook 安装 <!-- id: 17 -->
    - 创建 `cmd/codei18n/hook.go`
    - 实现 `hook install` 命令: 复制脚本到 `.git/hooks/pre-commit`
    - 编写 hook 脚本: 检查 staged 文件 -> 确保注释为英文 -> 自动更新 Mapping

## 阶段 7: 收尾与验证 (Wrap Up & Verification)

- [ ] **W001**: 编写集成测试套件 <!-- id: 18 -->
    - 创建 `tests/integration_test.go`
    - 模拟完整流程: Init -> Scan -> Translate -> Map Update -> IDE Output
- [ ] **W002**: 验证 "Pure JSON" 输出 <!-- id: 19 -->
    - 编写自动化脚本，确保 CLI 在 pipe 模式下输出的 JSON 可被 `jq` 解析
- [ ] **W003**: 代码覆盖率检查 <!-- id: 20 -->
    - 运行 `go test -coverprofile=coverage.out ./...`
    - 确保核心模块覆盖率 >= 80%
