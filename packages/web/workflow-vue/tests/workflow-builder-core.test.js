import { describe, it, expect } from "vitest";

// 简化的工作流构建器核心功能
function createWorkflowBuilder() {
  const defaultWorkflow = [
    {
      NodeCode: "start",
      NodeName: "Submit",
      NodeType: "START",
      SortOrder: 0,
      ApproverConfig: "{}",
    },
    {
      NodeCode: "approval",
      NodeName: "Manager Approval",
      NodeType: "APPROVAL",
      SortOrder: 1,
      ApproverConfig: '{"type":"USERS","users":[],"mode":"OR"}',
    },
    {
      NodeCode: "end",
      NodeName: "End",
      NodeType: "END",
      SortOrder: 2,
      ApproverConfig: "{}",
    },
  ];

  return {
    workflowData: [...defaultWorkflow],

    loadFromBackendData(data) {
      if (!Array.isArray(data)) {
        return { success: true, data: defaultWorkflow };
      }
      this.workflowData = data;
      return { success: true, data };
    },

    getBackendData() {
      return this.workflowData.map((node, index) => ({
        ...node,
        SortOrder: index,
      }));
    },

    addNode(node) {
      const newNode = {
        ...node,
        SortOrder: this.workflowData.length,
      };
      this.workflowData.push(newNode);
      return newNode;
    },

    deleteNode(index) {
      if (index >= 0 && index < this.workflowData.length) {
        this.workflowData.splice(index, 1);
        // 重新排序
        this.workflowData.forEach((node, i) => {
          node.SortOrder = i;
        });
      }
    },
  };
}

describe("WorkflowBuilder Core", () => {
  it("loads backend data successfully", () => {
    const builder = createWorkflowBuilder();
    const customData = [
      { NodeName: "Custom Start", NodeType: "START" },
      { NodeName: "Custom End", NodeType: "END" },
    ];

    const result = builder.loadFromBackendData(customData);
    expect(result.success).toBe(true);
    expect(result.data).toBe(customData);
  });

  it("uses default data for invalid backend data", () => {
    const builder = createWorkflowBuilder();
    const result = builder.loadFromBackendData(null);

    expect(result.success).toBe(true);
    expect(result.data).toHaveLength(3);
    expect(result.data[0].NodeType).toBe("START");
  });

  it("reorders nodes correctly", () => {
    const builder = createWorkflowBuilder();
    const result = builder.getBackendData();

    expect(result).toHaveLength(3);
    result.forEach((node, index) => {
      expect(node.SortOrder).toBe(index);
    });
  });

  it("adds new node to workflow", () => {
    const builder = createWorkflowBuilder();
    const newNode = builder.addNode({
      NodeName: "Review",
      NodeType: "APPROVAL",
    });

    expect(builder.workflowData).toHaveLength(4);
    expect(newNode.SortOrder).toBe(3);
    expect(newNode.NodeName).toBe("Review");
  });

  it("deletes node and reorders", () => {
    const builder = createWorkflowBuilder();
    builder.deleteNode(1); // Delete approval node

    expect(builder.workflowData).toHaveLength(2);
    expect(builder.workflowData[0].NodeType).toBe("START");
    expect(builder.workflowData[1].NodeType).toBe("END");
    expect(builder.workflowData[1].SortOrder).toBe(1);
  });
});
