# 技术实现计划: 批量翻译优化

**功能**: 批量翻译优化 (`batch-translation`)
**相关规范**: [specs/003-batch-translation/spec.md](spec.md)
**创建时间**: 2025-12-17
**状态**: 规划中

## 摘要

优化现有的 `TranslateBatch` 方法，使其真正利用 LLM 的上下文窗口，一次性处理多条翻译任务，而不是简单的循环调用。
核心逻辑是构建包含 JSON 数组的 Prompt，并解析 LLM 返回的 JSON 数组，同时具备长度检查和错误重试机制。

## 技术背景

**语言/版本**: Go 1.21+
**主要依赖**: `sashabaranov/go-openai`
**存储**: 无（状态在内存中）
**测试**: Go unit tests (Mock LLM Response)

## 章程检查

**I. AST 优先原则**: N/A (此功能属于翻译层，不涉及 AST)
**II. 单一语言源原则**: N/A
**III. 代码规范与测试**:
- [x] 单元测试覆盖率 ≥ 60%
- [x] 使用 gofmt 和 golint
**IV. 中文优先**:
- [x] 注释和文档使用中文
**V. CLI 优先**:
- [x] CLI `translate` 命令将自动受益于此优化

## 项目结构

### 源代码

```
adapters/
└── translator/
    └── llm.go            # 修改: 实现真正的 TranslateBatch
    └── llm_test.go       # 新增/修改: 测试 Batch 逻辑
```

## 架构设计

### Prompt 设计

为了保证稳定性，使用 JSON Array 格式。

**Request Prompt**:
```text
You are a code comment translator. Translate the following JSON array of comments from {SourceLang} to {TargetLang}.

Rules:
1. Maintain the JSON array format.
2. The output must be a valid JSON string array ["...","..."].
3. The number of elements MUST match the input.
4. Keep code/variables unchanged.
5. Translate comments to {TargetLang}.

Input:
[
  "comment 1",
  "comment 2",
  ...
]
```

**Response Expected**:
```json
[
  "翻译 1",
  "翻译 2",
  ...
]
```

### 错误处理与回退 (Fallback)

1.  **JSON 解析失败**: 如果 LLM 返回非 JSON 格式（如包含Markdown代码块标记），先尝试清理（去除 \`\`\`json），再解析。如果仍失败，回退到逐条翻译。
2.  **长度不匹配**: 如果 `len(output) != len(input)`，视为失败，回退到逐条翻译（因为无法确定哪条对应哪条）。

### 实现细节

在 `LLMTranslator.TranslateBatch` 中：
1.  如果 `len(texts) <= 1`，直接调用 `Translate`。
2.  构造 Batch Prompt。
3.  调用 API。
4.  解析 Response。
5.  验证长度。
6.  如果成功，返回结果。
7.  如果任何步骤失败，记录 Warning，并降级为循环调用 `Translate` (Sequential Fallback)。

## 复杂度跟踪

| 违规 | 为什么需要 | 拒绝更简单替代方案的原因 |
|-----------|------------|-------------------------------------|
| Batching Logic | 性能与成本 | 逐条翻译对于大项目（1000+条注释）太慢，且容易触发 API Rate Limit (429)。Batching 是必要的优化。 |
