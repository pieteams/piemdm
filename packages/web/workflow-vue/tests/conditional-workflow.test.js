import { describe, it, expect } from "vitest";

// 条件分支工作流处理
function processConditionalWorkflow(nodes) {
  if (!Array.isArray(nodes)) {
    return { valid: false, data: [], error: "Invalid workflow data" };
  }

  const result = [];
  const branchNodes = new Map();

  // 第一阶段：收集所有分支节点
  nodes.forEach(node => {
    if (node.NodeType === "CONDITION") {
      const config = JSON.parse(node.ConditionConfig || "{}");
      config.branches?.forEach(branch => {
        branch.nodes?.forEach(branchNode => {
          const code = branchNode.nodeCode || branchNode.NodeCode;
          if (code) {
            branchNodes.set(code, branchNode);
          }
        });
      });
    }
  });

  // 第二阶段：处理主节点，跳过已在分支中的节点
  let sortOrder = 0;
  nodes.forEach(node => {
    if (node.NodeType !== "CONDITION" || !branchNodes.has(node.NodeCode)) {
      result.push({
        ...node,
        SortOrder: sortOrder++
      });
    }

    // 添加条件分支节点
    if (node.NodeType === "CONDITION") {
      const config = JSON.parse(node.ConditionConfig || "{}");
      config.branches?.forEach(branch => {
        branch.nodes?.forEach(branchNode => {
          result.push({
            NodeCode: branchNode.nodeCode || branchNode.NodeCode,
            NodeName: branchNode.nodeName || branchNode.NodeName,
            NodeType: branchNode.nodeType || branchNode.NodeType,
            SortOrder: sortOrder++
          });
        });
      });
    }
  });

  return { valid: true, data: result };
}

describe("Conditional Workflow", () => {
  it("processes workflow with simple condition", () => {
    const workflow = [
      {
        NodeCode: "start",
        NodeName: "Submit",
        NodeType: "START"
      },
      {
        NodeCode: "condition",
        NodeName: "Check Amount",
        NodeType: "CONDITION",
        ConditionConfig: JSON.stringify({
          branches: [
            {
              name: "High Amount",
              nodes: [
                {
                  nodeCode: "high-approval",
                  nodeName: "High Amount Approval",
                  nodeType: "APPROVAL"
                }
              ]
            }
          ]
        })
      },
      {
        NodeCode: "end",
        NodeName: "End",
        NodeType: "END"
      }
    ];

    const result = processConditionalWorkflow(workflow);

    expect(result.valid).toBe(true);
    expect(result.data).toHaveLength(4);
    expect(result.data[2].NodeCode).toBe("high-approval");
    expect(result.data[2].NodeName).toBe("High Amount Approval");
  });

  it("handles multiple branches", () => {
    const workflow = [
      {
        NodeCode: "condition",
        NodeType: "CONDITION",
        ConditionConfig: JSON.stringify({
          branches: [
            {
              nodes: [
                { nodeCode: "branch-1", nodeName: "Branch One", nodeType: "APPROVAL" }
              ]
            },
            {
              nodes: [
                { nodeCode: "branch-2", nodeName: "Branch Two", nodeType: "APPROVAL", approverType: "AUTO_REJECT" }
              ]
            }
          ]
        })
      }
    ];

    const result = processConditionalWorkflow(workflow);
    expect(result.data).toHaveLength(3);
    expect(result.data[1].NodeName).toBe("Branch One");
    expect(result.data[2].NodeName).toBe("Branch Two");
  });

  it("handles invalid workflow data", () => {
    const result = processConditionalWorkflow(null);
    expect(result.valid).toBe(false);
    expect(result.error).toBe("Invalid workflow data");
  });

  it("preserves node properties in branches", () => {
    const workflow = [{
      NodeType: "CONDITION",
      ConditionConfig: JSON.stringify({
        branches: [{
          nodes: [{
            nodeCode: "test-branch",
            nodeName: "Test Branch",
            nodeType: "APPROVAL",
            approverConfig: '{"type":"USERS","users":["user1"]}'
          }]
        }]
      })
    }];

    const result = processConditionalWorkflow(workflow);
    expect(result.data[1].NodeCode).toBe("test-branch");
    expect(result.data[1].NodeType).toBe("APPROVAL");
  });

  it("handles empty branches gracefully", () => {
    const workflow = [{
      NodeType: "CONDITION",
      ConditionConfig: JSON.stringify({ branches: [] })
    }];

    const result = processConditionalWorkflow(workflow);
    expect(result.valid).toBe(true);
    expect(result.data).toHaveLength(1);
  });
});
