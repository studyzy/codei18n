# 规范增量：JavaScript 与 TypeScript 适配器

## 新增需求

### 需求：文件扩展名支持

适配器必须支持以下文件扩展名：
- `.js` (JavaScript)
- `.jsx` (JavaScript with JSX)
- `.ts` (TypeScript)
- `.tsx` (TypeScript with JSX)

#### 场景：识别扩展名
给定文件 `test.ts`，适配器应能正确加载 TypeScript 解析器。

### 需求：注释类型识别

适配器必须能够识别并区分以下注释类型：

#### 场景：行注释
对于以 `//` 开头的注释：
- 类型应识别为 `CommentTypeLine`。
- 内容应去除开头的 `//` 及紧随的空白。

#### 场景：块注释
对于以 `/*` 开头并以 `*/` 结尾的注释（非文档注释）：
- 类型应识别为 `CommentTypeBlock`。
- 内容应去除定界符 `/*`, `*/`。

#### 场景：文档注释 (JSDoc/TSDoc)
对于以 `/**` 开头的注释：
- 类型应识别为 `CommentTypeDoc`。
- 必须保留其作为文档注释的语义属性。

### 需求：符号绑定 (Symbol Binding)

注释必须尽可能绑定到紧随其后的代码符号（Symbol）。

#### 场景：函数声明
```typescript
// 这是一个函数
function calculateSum(a, b) { ... }
```
- Symbol: `calculateSum`

#### 场景：类方法
```typescript
class Calculator {
    // 计算方法
    add(a, b) { ... }
}
```
- Symbol: `Calculator.add`

#### 场景：箭头函数赋值
```typescript
// 处理器
const handler = () => { ... }
```
- Symbol: `handler`

#### 场景：TypeScript 接口
```typescript
// 用户接口
interface User {
    name: string;
}
```
- Symbol: `User`

#### 场景：顶级语句
如果注释不属于任何特定定义（如文件顶部的版权声明），Symbol 可为空或标识为 `file` 级别。

### 需求：健壮性

适配器必须具有容错能力。

#### 场景：语法错误
对于包含语法错误的源文件，适配器应尽力提取可识别的注释，不应导致程序崩溃 (Panic)。
