# 变更：重构翻译后端为 LLM + 本地 Ollama

## 为什么
当前翻译后端依赖 Google / DeepL 等传统 API，存在隐私与合规风险，也缺乏对本地 LLM 的支持。
我们希望统一到 OpenAI 兼容协议的 LLM 接口，并支持本地 Ollama，以满足离线和隐私敏感场景。

## 变更内容
- 移除对 Google / DeepL 作为 `translator.provider` 的支持。
- 引入统一的翻译配置模型，`translator.provider` 仅支持：`llm-api`、`ollama`。
- 默认 `translator.provider` 为 `llm-api`，基于 OpenAI 兼容协议（可配置 `baseUrl`、`model` 与 `apiKeyEnv`）。
- 新增 `OllamaTranslator`，通过本地 Ollama 服务执行翻译，失败时不自动 fallback 到云端。
- 引入统一的 `TranslatorConfig` 和 `NewTranslator(cfg)` 工厂方法，集中选择后端实现。
- 更新 README / CODEBUDDY / 示例配置，明确当前官方支持的后端矩阵，并标记 Google / DeepL 为历史方案。

## 影响
- 受影响规范：翻译引擎能力、配置文件语义（翻译后端选择、默认行为、本地模式）。
- 受影响代码：
  - core/translate/...（Translator 实现与工厂）
  - core/mapping/...（如果依赖后端行为，需要确认）
  - CLI / pre-commit hook 中 Translator 初始化逻辑
- 受影响用户：
  - 所有仍在使用 `google` / `deepl` 作为 `translator.provider` 的配置将不再工作，需要手动迁移到 `llm-api` 或 `ollama`。
