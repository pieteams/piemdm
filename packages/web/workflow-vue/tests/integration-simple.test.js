import { describe, it, expect } from "vitest";

// 简化的工作流集成测试
function createWorkflow(nodes) {
  return {
    nodes: nodes || [],
    hasNodes() {
      return this.nodes.length > 0;
    },
    getStartNode() {
      return this.nodes.find(node => node.NodeType === "START");
    },
    getEndNode() {
      return this.nodes.find(node => node.NodeType === "END");
    }
  };
}

describe("Workflow Integration", () => {
  it("creates workflow with start and end nodes", () => {
    const workflow = createWorkflow([
      { NodeName: "Start", NodeType: "START" },
      { NodeName: "End", NodeType: "END" }
    ]);

    expect(workflow.hasNodes()).toBe(true);
    expect(workflow.getStartNode().NodeName).toBe("Start");
    expect(workflow.getEndNode().NodeName).toBe("End");
  });

  it("handles empty workflow", () => {
    const workflow = createWorkflow();
    expect(workflow.hasNodes()).toBe(false);
    expect(workflow.getStartNode()).toBeUndefined();
    expect(workflow.getEndNode()).toBeUndefined();
  });

  it("processes approval flow", () => {
    const workflow = createWorkflow([
      { NodeName: "Submit", NodeType: "START" },
      { NodeName: "Manager", NodeType: "APPROVAL" },
      { NodeName: "Complete", NodeType: "END" }
    ]);

    const approver = workflow.nodes.find(node => node.NodeType === "APPROVAL");
    expect(approver.NodeName).toBe("Manager");
  });
});
