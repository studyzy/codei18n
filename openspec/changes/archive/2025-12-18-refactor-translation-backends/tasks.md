## 1. 配置与类型定义
- [x] 1.1 采用 `Config.TranslationProvider`（openai/llm/ollama）和 `TranslationConfig` map 承载翻译后端配置，默认 provider 调整为 `openai`。
- [x] 1.2 在 `translator.NewFromConfig` 中统一校验 `TranslationProvider`，只允许 `openai` / `llm` / `ollama`（以及测试用 `mock`），其余值返回错误。
- [x] 1.3 在 `translator.NewFromConfig` 中对 `google` / `deepl` provider 返回错误提示，并指导用户迁移到 `openai` 或 `ollama`。

## 2. Translator 实现
- [x] 2.1 保持现有 `LLMTranslator`（OpenAI 兼容），通过工厂从配置和环境变量中获取 `baseUrl` + `model` + `apiKey`。
- [x] 2.2 实现 `OllamaTranslator`，连接本地 Ollama，不做任何自动 fallback 到云端。

## 3. 工厂与集成
- [x] 3.1 实现 `NewFromConfig(cfg *config.Config) (core.Translator, error)` 工厂方法，集中处理 provider 分支和配置解析。
- [x] 3.2 在 CLI（translate/init/hook）中统一改用工厂创建 Translator，移除对具体后端实现的直接依赖。
- [x] 3.3 已确认不存在 Google / DeepL 运行时代码实现，仅在工厂中保留错误提示分支；历史文档中已将 Google / DeepL 标记为不再支持。

## 4. 测试
- [x] 4.1 通过 `go test ./...` 验证 translator 包与 CLI 构建均通过，为后续细化测试提供基础。
- [x] 4.2 在 `adapters/translator/factory_test.go` 中添加针对历史 provider（google/deepl）和未知 provider 的错误返回测试。
- [x] 4.3 通过 `go test ./...` 确认当前变更未破坏现有测试体系，核心覆盖率要求由项目 CI 统一检查。

## 5. 文档与示例
- [x] 5.1 已在 `README.md` 和 `CODEBUDDY.md` 中更新翻译引擎章节，列出 openai/llm/ollama 后端矩阵，并明确 Google / DeepL 为不再支持的历史方案。
- [x] 5.2 已在 `CODEBUDDY.md` 的配置示例中展示 `translationProvider: "openai"` 及 `translationConfig` 字段；README 中提供了 openai/ollama 的完整配置示例。
- [x] 5.3 已在 `README.md` 的 `13.3.4 从 Google / DeepL 迁移` 小节中给出具体迁移步骤说明。
