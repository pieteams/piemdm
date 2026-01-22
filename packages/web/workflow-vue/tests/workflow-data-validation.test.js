import { describe, it, expect } from "vitest";

// 简化的工作流数据验证函数
function validateWorkflowData(nodes) {
  if (!Array.isArray(nodes)) {
    return { valid: false, error: "Data must be an array" };
  }

  if (nodes.length === 0) {
    return { valid: false, error: "Workflow cannot be empty" };
  }

  // 检查SortOrder连续性要求
  const hasSortOrder = nodes.some((node) => node.SortOrder !== undefined);
  if (hasSortOrder) {
    const sortOrders = nodes.map((n) => n.SortOrder || 0).sort((a, b) => a - b);
    for (let i = 0; i < sortOrders.length; i++) {
      if (sortOrders[i] !== i) {
        return { valid: false, error: "SortOrder must be continuous" };
      }
    }
  }

  // 检查节点类型有效性
  const validNodeTypes = ["START", "END", "APPROVAL", "CONDITION", "CC"];
  const requiredFields = ["NodeCode", "NodeName", "NodeType"];
  for (const node of nodes) {
    // 检查必需字段
    for (const field of requiredFields) {
      if (!node[field]) {
        return { valid: false, error: `Missing required field: ${field}` };
      }
    }

    // 检查节点类型
    if (!validNodeTypes.includes(node.NodeType)) {
      return { valid: false, error: `Invalid NodeType: ${node.NodeType}` };
    }
  }

  return { valid: true };
}

describe("Workflow Data Validation", () => {
  it("validates missing required fields", () => {
    const invalidData = [{ NodeCode: "test" }]; // Missing NodeName, NodeType, SortOrder
    const result = validateWorkflowData(invalidData);
    expect(result.valid).toBe(false);
    expect(result.error).toContain("Missing required field");
  });

  it("validates invalid node type", () => {
    const invalidData = [
      {
        NodeCode: "test",
        NodeName: "Test",
        NodeType: "INVALID_TYPE",
        SortOrder: 0,
      },
    ];
    const result = validateWorkflowData(invalidData);
    expect(result.valid).toBe(false);
    expect(result.error).toContain("Invalid NodeType");
  });

  it("validates proper workflow data", () => {
    const validData = [
      { NodeCode: "start", NodeName: "Start", NodeType: "START" },
      { NodeCode: "end", NodeName: "End", NodeType: "END" },
    ];
    const result = validateWorkflowData(validData);
    expect(result.valid).toBe(true);
  });

  it("validates empty workflow", () => {
    const result = validateWorkflowData([]);
    expect(result.valid).toBe(false);
    expect(result.error).toBe("Workflow cannot be empty");
  });

  it("validates non-array data", () => {
    const result = validateWorkflowData(null);
    expect(result.valid).toBe(false);
    expect(result.error).toBe("Data must be an array");
  });
});
