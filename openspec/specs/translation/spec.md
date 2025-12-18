# translation Specification

## Purpose
TBD - created by archiving change refactor-translation-backends. Update Purpose after archive.
## 需求
### 需求：翻译后端选择与默认行为
系统必须提供统一的翻译后端选择机制，当前仅支持以下两种后端：
- `llm-api`：通过 OpenAI 兼容协议访问云端或代理 LLM；
- `ollama`：通过本地 Ollama 实例执行翻译。

#### 场景：默认使用 llm-api
- **当** 用户未在配置中显式设置 `translator.provider`
- **那么** 系统必须默认使用 `llm-api` 作为翻译后端
- **并且** 当 `llm-api` 配置无效或鉴权失败时，系统必须返回清晰错误，而不是静默降级或自动切换到其他后端。

#### 场景：本地 Ollama 不自动降级
- **当** 用户在配置中将 `translator.provider` 设置为 `ollama`
- **并且** 本地 Ollama 服务不可用、连接失败或指定模型缺失
- **那么** 系统必须返回清晰错误，提示本地服务/模型不可用
- **并且** 系统不得在用户未显式配置的情况下自动调用任何云端翻译后端。

#### 场景：拒绝历史 provider
- **当** 配置中的 `translator.provider` 值为 `google` 或 `deepl`
- **那么** 系统必须在启动或执行 CLI 时立即报错
- **并且** 错误信息中必须明确提示用户迁移到 `llm-api` 或 `ollama`。

