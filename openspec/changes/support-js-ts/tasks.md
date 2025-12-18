# 任务列表：JS/TS 支持实施

- [x] <!-- id: 0 --> **依赖引入**
    - 更新 `go.mod` 引入 `tree-sitter-javascript` 和 `tree-sitter-typescript` bindings。
    - 验证依赖下载成功。

- [x] <!-- id: 1 --> **创建适配器骨架**
    - 创建 `adapters/typescript` 目录。
    - 实现基础结构 `adapter.go`，满足 `LanguageAdapter` 接口。
    - 实现根据文件后缀 (`.js`, `.jsx`, `.ts`, `.tsx`) 加载不同 Tree-sitter Grammar 的逻辑。

- [x] <!-- id: 2 --> **实现符号解析逻辑 (Symbol Resolution)**
    - 实现 `symbol.go`。
    - 编写逻辑处理函数声明 (`FunctionDeclaration`)。
    - 编写逻辑处理类与方法 (`ClassDeclaration`, `MethodDefinition`)。
    - 编写逻辑处理变量声明中的箭头函数 (`VariableDeclarator` -> `ArrowFunction`)。
    - 编写逻辑处理 TypeScript 特有结构 (`InterfaceDeclaration`, `TypeAliasDeclaration`, `EnumDeclaration`)。

- [x] <!-- id: 3 --> **实现注释提取与查询**
    - 编写 Tree-sitter 查询语句 (`queries.go`) 匹配行注释和块注释。
    - 处理 JSX/TSX 中的注释 `{/* ... */}`。
    - 确保文档注释 (JSDoc/TSDoc) 被正确标记为 `CommentTypeDoc`。

- [x] <!-- id: 4 --> **注册适配器**
    - 修改 `adapters/registry.go`，将 `.js`, `.jsx`, `.ts`, `.tsx` 映射到新适配器。

- [x] <!-- id: 5 --> **单元测试**
    - 编写 `adapters/typescript/adapter_test.go`。
    - 覆盖常见 JS/TS 语法场景。
    - 验证 Symbol 生成的正确性和稳定性。

- [x] <!-- id: 6 --> **集成测试**
    - 添加 JS/TS 样本文件到 `tests/testdata/`。
    - 更新 `tests/integration_test.go` 或新增 `js_ts_integration_test.go` 进行端到端验证。
