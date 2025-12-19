# 任务列表：Java 支持实施

- [x] <!-- id: 0 --> **依赖引入**
  - 在 `go.mod` 中引入 `github.com/smacker/go-tree-sitter/java` 依赖。
  - 确认依赖可以在本地和 CI 环境中正常下载与构建。

- [x] <!-- id: 1 --> **创建适配器骨架**
  - 在 `adapters` 目录下创建 `java` 子目录。
  - 在 `adapters/java/adapter.go` 中实现基础结构，满足 `LanguageAdapter` 接口。
  - 在 `adapters/java/queries.go` 中定义 Tree-sitter 查询，用于匹配 Java 注释节点。

- [x] <!-- id: 2 --> **实现符号解析逻辑 (Symbol Resolution)**
  - 在 `adapters/java/symbol.go` 中实现符号路径解析：
    - 顶级类/接口/枚举：`package.ClassName`；
    - 成员方法：`package.ClassName#methodName`；
    - 字段：`package.ClassName#fieldName`；
    - 内部类：`package.OuterClass$InnerClass`。
  - 处理文件级注释的 Symbol 表示（文件级或空 Symbol）。

- [x] <!-- id: 3 --> **实现注释提取与类型识别**
  - 在 `adapter.go` 中集成 Tree-sitter 解析与查询执行逻辑。
  - 正确区分并设置 `CommentTypeLine`、`CommentTypeBlock` 和 `CommentTypeDoc`。
  - 对注释内容去除前缀/后缀与多余空白，保证生成的注释文本适合作为 ID 计算输入。

- [x] <!-- id: 4 --> **注册适配器**
  - 修改 `adapters/registry.go`，为 `.java` 文件扩展名映射 Java 适配器。
  - 确认对其他语言（Go/Rust/JS/TS）的行为不产生回归。

- [x] <!-- id: 5 --> **单元测试**
  - 编写 `adapters/java/adapter_test.go`，覆盖：
    - 类、接口、枚举、内部类、字段、方法等常见结构；
    - Javadoc、块注释、行注释的识别与类型映射；
    - 符号路径生成的稳定性验证。

- [x] <!-- id: 6 --> **集成测试**
  - 在 `tests/testdata/` 下新增若干 Java 样例文件，覆盖常见和边界语法场景。
  - 在现有集成测试中增加针对 `.java` 文件的扫描用例，验证 CLI `codei18n scan` 可以正确输出 Java 注释映射。
