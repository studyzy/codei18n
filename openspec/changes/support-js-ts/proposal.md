# 提案：增加 JavaScript 和 TypeScript 支持

## 变更摘要
本提案旨在为 CodeI18n 增加对 JavaScript (.js, .jsx) 和 TypeScript (.ts, .tsx) 语言的支持。将通过集成 `tree-sitter-javascript` 和 `tree-sitter-typescript` 语法解析器，实现对这两类语言源代码中注释的提取、定位和还原，从而支持其国际化处理。

## 动机
JavaScript 和 TypeScript 是现代 Web 开发的主流语言，拥有庞大的开发者群体。支持 JS/TS 将显著扩大 CodeI18n 的适用范围，使其能够服务于前端项目、Node.js 后端项目以及全栈应用。目前的架构已通过 Rust 适配器验证了基于 Tree-sitter 的多语言扩展能力，扩展 JS/TS 支持是顺理成章的下一步。

## 范围
- **新增适配器**：实现 JavaScript 和 TypeScript 的 `LanguageAdapter`。
- **文件扩展名支持**：支持 `.js`, `.jsx`, `.ts`, `.tsx`。
- **注释类型支持**：
  - 单行注释 `//`
  - 块注释 `/* */`
  - 文档注释 `/** */` (JSDoc/TSDoc)
- **依赖库**：复用现有的 `github.com/smacker/go-tree-sitter` 及其对应的 JS/TS bindings。
- **测试**：包含针对各种 JS/TS 语法结构的单元测试和集成测试。

## 风险
- **语法复杂性**：JS/TS 语法灵活（如 JSX, 装饰器, 复杂的类型定义），需要确保注释与其所属的语义节点（Symbol）准确绑定。
- **性能**：Tree-sitter 解析大文件时的性能需关注，但鉴于 Rust 适配器已验证其可行性，风险可控。
