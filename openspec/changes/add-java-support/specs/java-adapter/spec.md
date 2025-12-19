# 规范增量：Java 适配器

## 新增需求

### 需求：文件扩展名支持

适配器必须支持对 `.java` 源文件进行解析，并将其纳入 CodeI18n 的注释扫描范围。

#### 场景：识别 .java 扩展名
给定文件 `src/com/example/App.java`：
- **当** 通过 `codei18n scan --file src/com/example/App.java` 执行扫描时；
- **那么** 系统必须为该文件选择 Java 适配器进行解析；
- **并且** 不得返回“不支持的文件类型”错误。

### 需求：注释类型识别

Java 适配器必须能够识别并区分以下注释类型，并映射到统一注释模型中的类型枚举：
- 行注释 `//` → `CommentTypeLine`；
- 块注释 `/* ... */` → `CommentTypeBlock`；
- 文档注释 `/** ... */`（Javadoc）→ `CommentTypeDoc`。

#### 场景：行注释
给定如下代码片段：
```java
// 这是一个行注释
int value = 42;
```
- **当** 适配器解析该文件时；
- **那么** 应产生一个注释，其类型为 `CommentTypeLine`；
- **并且** 注释内容必须去除前缀 `//` 及紧随其后的空白字符。

#### 场景：块注释
给定如下代码片段：
```java
/* 这是一个块注释 */
int value = 42;
```
- **当** 适配器解析该文件时；
- **那么** 应产生一个注释，其类型为 `CommentTypeBlock`；
- **并且** 注释内容必须去除定界符 `/*` 与 `*/` 以及首尾多余空白。

#### 场景：文档注释 (Javadoc)
给定如下代码片段：
```java
/**
 * 计算两个数的和
 */
public int sum(int a, int b) { ... }
```
- **当** 适配器解析该文件时；
- **那么** 应产生一个注释，其类型为 `CommentTypeDoc`；
- **并且** 必须保留其作为文档注释的语义属性，用于后续在 IDE 中以不同样式渲染。

### 需求：符号绑定 (Symbol Binding)

Java 适配器必须尽可能将注释绑定到紧随其后的代码符号（Symbol），并为常见结构生成稳定且可读的符号路径。

#### 场景：类声明
```java
// 计算器类
public class Calculator { }
```
- **Symbol** 应解析为：`package.Calculator`（其中 `package` 为文件声明的包名；如果无包声明则只使用 `Calculator`）。

#### 场景：成员方法
```java
public class Calculator {
    // 计算两个数的和
    public int add(int a, int b) { ... }
}
```
- **Symbol** 应解析为：`package.Calculator#add`。

#### 场景：字段
```java
public class Config {
    // 默认超时时间（毫秒）
    private int timeoutMs = 1000;
}
```
- **Symbol** 应解析为：`package.Config#timeoutMs`。

#### 场景：接口与枚举
```java
// 用户接口
public interface User { }

// 账户类型
public enum AccountType {
    STANDARD,
    PREMIUM
}
```
- 接口注释的 **Symbol** 应解析为：`package.User`；
- 枚举注释的 **Symbol** 应解析为：`package.AccountType`。

#### 场景：文件级注释
```java
// Copyright (c) 2025 Example Corp
// 本文件包含示例代码
package com.example;
```
- 当注释位于文件顶部且不紧邻任何具体类型/成员声明时：
  - 适配器可以将 Symbol 视为文件级别标识（例如仅使用包名或空 Symbol），但必须保证不会与后续类/方法的注释混淆。

### 需求：健壮性

Java 适配器在面对语法错误或不完整代码时必须具备容错能力，尽可能返回可用的注释信息，而不是导致程序崩溃。

#### 场景：语法错误文件
给定如下包含语法错误的代码：
```java
public class Broken {
    // 未完成的方法定义
    public void doSomething( {
}
```
- **当** 使用 Java 适配器解析该文件时；
- **那么** 适配器必须至少返回注释 `// 未完成的方法定义` 对应的注释实体；
- **并且** 不得触发 panic 或导致整个扫描流程中止；
- **如果** 语法错误严重到无法构建任何语法树，适配器可以返回空注释列表，但仍应通过 `error` 明确指出解析失败原因。
