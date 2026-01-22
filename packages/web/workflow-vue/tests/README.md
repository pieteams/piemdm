# 工作流组件测试文档

## vtest 标准格式

本项目采用 vtest 标准格式编写测试用例，遵循以下原则：

### 核心原则
- **精简（Minimal）**：每个测试文件专注于单一职责
- **专注（Focused）**：测试用例清晰、有针对性
- **可维护（Maintainable）**：代码结构清晰，易于理解和修改
- **快速（Fast）**：测试执行效率高

### 测试文件结构
- `uuid-utils.test.js` - UUID工具函数测试
- `condition-branches.test.js` - 条件分支处理测试
- `integration-simple.test.js` - 工作流集成测试
- `workflow-data-validation.test.js` - 数据验证测试
- `workflow-builder-core.test.js` - 核心工作流构建器测试
- `conditional-workflow.test.js` - 条件工作流测试

### 编写规范
1. 使用英文描述测试用例
2. 每个测试函数只测试一个具体功能点
3. 避免复杂的Mock类，使用简单函数
4. 测试数据保持最简
5. 断言清晰明确

### 示例代码
```javascript
describe("UUID Utils", () => {
  it("generates valid UUID format", () => {
    const uuid = uuidv4();
    expect(uuid).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i);
    expect(uuid).toHaveLength(36);
  });
});
```

## 运行测试
```bash
# 运行所有测试
pnpm test

# 运行特定测试文件
pnpm test --run uuid-utils.test.js

# 运行vtest标准测试
pnpm test --run uuid-utils.test.js condition-branches.test.js integration-simple.test.js workflow-data-validation.test.js workflow-builder-core.test.js conditional-workflow.test.js
```
