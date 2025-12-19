# 设计文档：Java 适配器设计

## 上下文

- 项目已经存在 Go、Rust、JavaScript/TypeScript 三类语言的适配器：
  - Go：基于 `go/parser` 与 `go/ast` 的原生解析。
  - Rust：基于 `github.com/smacker/go-tree-sitter/rust` 的 Tree-sitter 适配器。
  - JS/TS：基于 `github.com/smacker/go-tree-sitter/{javascript,typescript,tsx}` 的 Tree-sitter 适配器，目录为 `adapters/typescript`。
- 现有架构通过 `core.LanguageAdapter` 接口与 `adapters/registry.go` 的扩展名分派机制统一管理多语言支持。
- JS/TS 适配器的设计文档已给出 Tree-sitter 适配器的推荐模式（`queries.go` + `symbol.go` + `adapter.go` + 测试）。

在此基础上，本设计文档给出 Java 适配器的具体设计方案。

## 目标 / 非目标

- 目标：
  - 为 `.java` 文件提供稳定的注释抽取与符号绑定能力。
  - 复用 Tree-sitter 适配器模式，尽量与 Rust/JS/TS 适配器在结构和测试策略上保持一致。
  - 不改变现有 CLI 使用方式，用户通过 `codei18n scan` 等命令即可透明获取 Java 支持。
- 非目标：
  - 不在本变更中支持 Kotlin/Scala 等其他 JVM 语言（可以在后续变更中复用本设计）。
  - 不修改翻译引擎与映射存储格式（只扩展可被扫描的语言范围）。

## 决策

1. **解析技术选择：Tree-sitter Java**
   - 方案 A：使用 Tree-sitter Java grammar（`github.com/smacker/go-tree-sitter/java`）。
   - 方案 B：使用 Java 编译器或 ANTLR 等传统解析器。
   - **决策**：选择方案 A，原因：
     - 与现有 Rust 和 JS/TS 适配器保持一致的依赖栈和解析模型；
     - Tree-sitter 在不完整/包含语法错误源码上的鲁棒性更好，适合增量开发场景；
     - go-tree-sitter 已提供 Java grammar 绑定，接入成本较低。

2. **目录结构与文件划分**
   - 在 `adapters/` 目录下新增 `java` 子目录：

     ```
     adapters/
       └── java/
           ├── adapter.go       # 实现 LanguageAdapter
           ├── queries.go       # Tree-sitter 查询定义
           ├── symbol.go        # 符号路径解析逻辑
           └── adapter_test.go  # 单元测试
     ```

   - 设计与 `adapters/typescript` 保持对齐：
     - `adapter.go` 负责构建 parser、执行解析与遍历注释节点；
     - `queries.go` 以 S-expression 定义注释节点匹配模式；
     - `symbol.go` 封装从注释节点解析语义符号路径的逻辑；
     - `adapter_test.go` 提供覆盖常见 Java 结构的测试样例。

3. **符号路径格式约定**
   - 参考项目章程中的示例：`Class#method` 用于表示面向对象语言的符号路径。
   - Java 符号路径建议格式：
     - 顶级类：`package.ClassName`
     - 成员方法：`package.ClassName#methodName`
     - 静态方法：`package.ClassName#methodName`（不额外区分 static，避免路径不稳定）；
     - 字段：`package.ClassName#fieldName`；
     - 接口：`package.InterfaceName`；
     - 枚举常量：`package.EnumName#CONSTANT`；
     - 内部类：`package.OuterClass$InnerClass`（与常见 JVM 命名习惯一致）。
   - 若无法确定 Symbol（例如文件头版权注释），则允许 Symbol 为空或使用文件级标识（由实现细节决定，不在本变更中强制约束具体值，只在规范中要求“可表示为文件级别”）。

4. **注释类型映射**
   - Java 注释语法：
     - 行注释：`// ...`
     - 块注释：`/* ... */`
     - 文档注释：`/** ... */`（Javadoc）
   - 与统一注释模型的映射：
     - `//` → `CommentTypeLine`；
     - `/* */` → `CommentTypeBlock`；
     - `/** */` → `CommentTypeDoc`。
   - 适配器需要在输出的 `Comment` 结构中正确设置 `Type` 字段，并对内容去除前缀/后缀与多余的空白字符。

5. **健壮性与错误处理**
   - Tree-sitter 能解析包含语法错误的源码，本适配器应复用这一能力：
     - 当出现语法错误或不完整代码时，适配器应尽可能提取可识别注释；
     - 禁止在解析阶段触发 panic，所有错误通过 `error` 返回；
     - 当 Tree-sitter 返回空语法树时，适配器应返回空注释列表而非错误，除非输入本身不可读（如 IO 错误）。

## 风险 / 权衡

- 风险 1：Tree-sitter Java 语法与实际使用的 JDK 版本差异可能导致部分新语法（如最新的 switch 表达式、record）暂时无法正确解析。
  - 缓解：
    - 在测试中加入常见现代语法样例，确认行为；
    - 对特殊节点保持“尽力而为”，即使无法完全理解结构，也尽量保留注释节点。
- 风险 2：复杂嵌套结构（如匿名内部类、Lambda 表达式）中的注释与符号绑定策略边界不清晰。
  - 缓解：
    - 在规范增量中给出典型场景示例，并在测试中覆盖；
    - 对过于模糊的场景允许回退为文件级 Symbol，以避免错误绑定。

## 迁移计划

1. 在不修改现有适配器行为的前提下，引入 Java 适配器代码与依赖；
2. 更新 `adapters/registry.go` 增加 `.java` 分派逻辑；
3. 增加针对 Java 的单元测试与集成测试，确保与现有语言并行执行；
4. 在文档与 README 中补充“支持 Java”的说明（可在后续文档类变更中通过单独提案完成）。

## 待决问题

- 是否需要在后续变更中引入统一的“language-adapters”能力规范，将 Go/Rust/JS/TS/Java 的行为抽象到同一规范下？
- 对于极端复杂的符号路径（如多重嵌套匿名类），是否需要在规范中进一步细化标识规则，或统一定义为“不可绑定到稳定 Symbol 时退回文件级别”？
