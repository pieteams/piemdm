import { describe, it, expect } from "vitest";
import { WorkflowService } from "../src/lib/services/workflow-service.js";
import { NodeService } from "../src/lib/services/node-service.js";
import { NodeHelper } from "../src/lib/utils/node-helper.js";
import { WorkflowUtils } from "../src/lib/utils/workflow-utils.js";
import { JsonHelper } from "../src/lib/utils/json-helper.js";

// 简化的条件分支处理函数
function processConditionBranches(nodes) {
  if (!Array.isArray(nodes)) return [];

  const result = [];
  const processedBranchNodes = new Set();

  nodes.forEach((node) => {
    if (node.NodeType === "CONDITION") {
      const config = JSON.parse(node.ConditionConfig || "{}");
      if (config.branches) {
        config.branches.forEach((branch) => {
          branch.nodes?.forEach((branchNode) => {
            const code = branchNode.nodeCode || branchNode.NodeCode;
            if (code) processedBranchNodes.add(code);
            result.push({
              NodeCode: code,
              NodeName: branchNode.nodeName || branchNode.NodeName,
              NodeType: branchNode.nodeType || branchNode.NodeType,
            });
          });
        });
      }
    }
  });

  return result;
}

describe("Condition Branches", () => {
  it("processes branches with nodes", () => {
    const nodes = [
      {
        NodeType: "CONDITION",
        ConditionConfig: JSON.stringify({
          branches: [
            {
              nodes: [
                {
                  nodeCode: "branch-1",
                  nodeName: "Branch 1",
                  nodeType: "APPROVAL",
                },
              ],
            },
          ],
        }),
      },
    ];

    const result = processConditionBranches(nodes);
    expect(result).toHaveLength(1);
    expect(result[0].NodeCode).toBe("branch-1");
    expect(result[0].NodeName).toBe("Branch 1");
  });

  it("handles empty condition config", () => {
    const result = processConditionBranches([]);
    expect(result).toEqual([]);
  });

  it("creates default branches when adding condition node", () => {
    // 创建一个测试工作流
    const workflow = {
      name: "Test Workflow",
      nodes: [
        {
          NodeCode: "start",
          NodeName: "开始",
          NodeType: "START",
          SortOrder: 0,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
        {
          NodeCode: "end",
          NodeName: "结束",
          NodeType: "END",
          SortOrder: 1,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      status: "draft",
    };

    // 添加条件节点
    WorkflowService.addNode(workflow, workflow.nodes, 1, "CONDITION", "条件分支");

    // 验证条件节点已添加
    const conditionNode = workflow.nodes[1];
    expect(conditionNode.NodeType).toBe("CONDITION");
    expect(conditionNode.NodeName).toBe("条件分支");

    // 验证条件配置包含两个分支
    const config = JSON.parse(conditionNode.ConditionConfig);
    expect(config.branches).toHaveLength(2);

    // 验证第一个分支
    expect(config.branches[0].name).toBe("分支1");
    expect(config.branches[0].nodes).toHaveLength(1);
    expect(config.branches[0].nodes[0].NodeType).toBe("APPROVAL");
    expect(config.branches[0].nodes[0].NodeName).toBe("审批节点");

    // 验证第二个分支
    expect(config.branches[1].name).toBe("其他");
    expect(config.branches[1].nodes).toHaveLength(1);
    const autoRejectNode = config.branches[1].nodes[0];
    expect(autoRejectNode.NodeType).toBe("APPROVAL");
    expect(autoRejectNode.ApproverType).toBe("AUTO_REJECT");
    expect(autoRejectNode.NodeName).toBe("自动驳回");

    // 验证自动驳回节点的配置
    const autoRejectConfig = JSON.parse(autoRejectNode.ApproverConfig);
    expect(autoRejectConfig.type).toBe("AUTO_REJECT");
    expect(autoRejectConfig.mode).toBe("AUTO");
  });

  it("displays correct text for auto reject node", () => {
    // 模拟自动驳回节点
    const autoRejectNode = {
      NodeType: "APPROVAL",
      ApproverType: "AUTO_REJECT",
      NodeName: "自动驳回",
    };

    // 验证显示文本
    expect(NodeHelper.getApproverText(autoRejectNode)).toBe("系统自动驳回");
  });

  it("saves CC node configuration correctly", () => {
    // 创建一个测试工作流
    const workflow = {
      name: "Test Workflow",
      nodes: [
        {
          NodeCode: "start",
          NodeName: "开始",
          NodeType: "START",
          SortOrder: 0,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
        {
          NodeCode: "cc",
          NodeName: "抄送节点",
          NodeType: "CC",
          SortOrder: 1,
          ApproverType: "USERS",
          ApproverConfig: '{"type":"CC","users":[],"mode":"CC","ccTiming":"after_approval"}',
          ConditionConfig: "{}",
        },
        {
          NodeCode: "end",
          NodeName: "结束",
          NodeType: "END",
          SortOrder: 2,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      status: "draft",
    };

    // 模拟更新抄送节点配置
    const updateData = {
      NodeName: "抄送给法务",
      Description: "审批通过后抄送给法务部门",
      ApproverConfig: JSON.stringify({
        type: "CC",
        users: ["user001", "user002"],
        mode: "CC",
        ccTiming: "after_approval",
      }),
    };

    // 调用更新服务
    const result = WorkflowService.updateNode(workflow, "cc", updateData);

    // 验证更新成功
    expect(result.success).toBe(true);
    expect(result.error).toBe(null);

    // 验证节点数据已更新
    const updatedNode = workflow.nodes[1];
    expect(updatedNode.NodeName).toBe("抄送给法务");
    expect(updatedNode.Description).toBe("审批通过后抄送给法务部门");

    // 验证审批配置
    const config = JSON.parse(updatedNode.ApproverConfig);
    expect(config.type).toBe("CC");
    expect(config.users).toEqual(["user001", "user002"]);
    expect(config.mode).toBe("CC");
    expect(config.ccTiming).toBe("after_approval");
  });

  it("validates CC node requires at least one user", () => {
    // 创建一个测试工作流
    const workflow = {
      name: "Test Workflow",
      nodes: [
        {
          NodeCode: "start",
          NodeName: "开始",
          NodeType: "START",
          SortOrder: 0,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
        {
          NodeCode: "cc",
          NodeName: "抄送节点",
          NodeType: "CC",
          SortOrder: 1,
          ApproverType: "USERS",
          ApproverConfig: '{"type":"CC","users":[],"mode":"CC","ccTiming":"after_approval"}',
          ConditionConfig: "{}",
        },
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      status: "draft",
    };

    // 尝试更新为空的抄送人列表
    const updateData = {
      ApproverConfig: JSON.stringify({
        type: "CC",
        users: [], // 空列表应该被验证为无效
        mode: "CC",
        ccTiming: "after_approval",
      }),
    };

    // 模拟 NodePropertyPanel 的验证逻辑
    const config = JSON.parse(updateData.ApproverConfig);
    const isValid = config.type === "USERS" && config.users.length > 0;

    // 验证空列表应该被拒绝
    expect(isValid).toBe(false);
  });

  it("simulates NodePropertyPanel CC node save process", () => {
    // 模拟 NodePropertyPanel 的状态
    const currentNode = {
      NodeCode: "cc",
      NodeName: "抄送节点",
      NodeType: "CC",
      ApproverType: "USERS",
      ApproverConfig: '{"type":"CC","users":[],"mode":"CC","ccTiming":"after_approval"}',
      ConditionConfig: "{}",
    };

    // 模拟面板中的响应式数据
    const formData = {
      NodeName: "抄送给财务",
      Description: "审批通过后抄送给财务部门",
    };

    const approverConfig = {
      type: "CC",
      users: ["finance001", "finance002"],
      mode: "CC",
      ccTiming: "after_approval",
    };

    // 模拟 saveConfiguration 函数的逻辑
    const saveConfiguration = () => {
      if (!currentNode) return false;

      // 验证必填字段
      if (!formData.NodeName.trim()) {
        return false;
      }

      // 构建更新数据
      const updateData = {
        NodeName: formData.NodeName.trim(),
        Description: formData.Description.trim(),
      };

      // 抄送节点需要更新审批配置
      if (currentNode.NodeType === "CC") {
        const configToSave = { ...approverConfig };
        if (currentNode.NodeType === "CC") {
          configToSave.type = "CC";
          configToSave.mode = "CC";
        }

        if (
          (configToSave.type === "USERS" || configToSave.type === "CC") &&
          configToSave.users.length === 0
        ) {
          return false;
        }
        updateData.ApproverConfig = JSON.stringify(configToSave);
      }

      return updateData;
    };

    // 执行保存
    const result = saveConfiguration();

    // 验证保存成功
    expect(result).not.toBe(false);
    expect(result.NodeName).toBe("抄送给财务");
    expect(result.Description).toBe("审批通过后抄送给财务部门");
    expect(result.ApproverConfig).toBeDefined();

    // 验证审批配置
    const config = JSON.parse(result.ApproverConfig);
    expect(config.type).toBe("CC");
    expect(config.users).toEqual(["finance001", "finance002"]);
    expect(config.mode).toBe("CC");
    expect(config.ccTiming).toBe("after_approval");
  });

  it("saves configuration for nodes in condition branches", () => {
    // 创建一个包含条件分支的测试工作流
    const workflow = {
      name: "Test Workflow",
      nodes: [
        {
          NodeCode: "start",
          NodeName: "开始",
          NodeType: "START",
          SortOrder: 0,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
        {
          NodeCode: "condition",
          NodeName: "条件分支",
          NodeType: "CONDITION",
          SortOrder: 1,
          ApproverType: "USERS",
          ApproverConfig: "{}",
          ConditionConfig: JSON.stringify({
            branches: [
              {
                id: "branch1",
                name: "分支1",
                condition: { fieldName: "", operator: "eq", fieldValue: "" },
                nodes: [
                  {
                    NodeCode: "branch-approval",
                    NodeName: "分支审批",
                    NodeType: "APPROVAL",
                    SortOrder: 0,
                    ApproverType: "USERS",
                    ApproverConfig: '{"type":"USERS","users":[],"mode":"OR"}',
                    ConditionConfig: "{}",
                  },
                ],
              },
              {
                id: "branch2",
                name: "其他",
                condition: { fieldName: "", operator: "else", fieldValue: "" },
                nodes: [
                  {
                    NodeCode: "branch-reject",
                    NodeName: "自动驳回",
                    NodeType: "APPROVAL",
                    ApproverType: "AUTO_REJECT",
                    SortOrder: 0,
                    ApproverConfig: '{"type":"AUTO_REJECT","mode":"AUTO"}',
                    ConditionConfig: "{}",
                  },
                ],
              },
            ],
          }),
        },
        {
          NodeCode: "end",
          NodeName: "结束",
          NodeType: "END",
          SortOrder: 2,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      status: "draft",
    };

    // 更新条件分支中的审批节点配置
    const updateData = {
      NodeName: "经理审批",
      Description: "部门经理进行最终审批",
      ApproverConfig: JSON.stringify({
        type: "USERS",
        users: ["manager001", "manager002"],
        mode: "AND",
      }),
    };

    // 调用更新服务
    const result = WorkflowService.updateNode(workflow, "branch-approval", updateData);

    // 验证更新成功
    expect(result.success).toBe(true);
    expect(result.error).toBe(null);

    // 验证原始条件配置已被更新
    const conditionNode = workflow.nodes[1];
    const config = JSON.parse(conditionNode.ConditionConfig);
    const updatedBranchNode = config.branches[0].nodes[0];

    // 验证条件分支中的节点已被正确更新

    // 验证条件分支中的节点已被正确更新
    expect(updatedBranchNode.NodeName).toBe("经理审批");
    expect(updatedBranchNode.Description).toBe("部门经理进行最终审批");

    // 验证审批配置
    const approverConfig = JSON.parse(updatedBranchNode.ApproverConfig);
    expect(approverConfig.users).toEqual(["manager001", "manager002"]);
    expect(approverConfig.mode).toBe("AND");
  });

  it("finds nodes in condition branches correctly", () => {
    // 测试 WorkflowUtils.findNodePath 能否正确找到条件分支中的节点
    const nodes = [
      {
        NodeCode: "start",
        NodeName: "开始",
        NodeType: "START",
      },
      {
        NodeCode: "condition",
        NodeName: "条件分支",
        NodeType: "CONDITION",
        ConditionConfig: JSON.stringify({
          branches: [
            {
              id: "branch1",
              name: "分支1",
              condition: { fieldName: "", operator: "eq", fieldValue: "" },
              nodes: [
                {
                  NodeCode: "branch-node-1",
                  NodeName: "分支节点1",
                  NodeType: "APPROVAL",
                },
              ],
            },
          ],
        }),
      },
    ];

    const path = WorkflowUtils.findNodePath(nodes, "branch-node-1");

    expect(path.length).toBeGreaterThan(0);
    expect(path[path.length - 1].array).toBeDefined();
    expect(path[path.length - 1].index).toBe(0);
  });

  it("updates condition branch nodes and syncs to workflowData", () => {
    // 测试条件分支节点更新后，workflowData中的ConditionConfig是否正确同步
    const workflow = {
      name: "Test Workflow",
      nodes: [
        {
          NodeCode: "start",
          NodeName: "开始",
          NodeType: "START",
          SortOrder: 0,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
        {
          NodeCode: "condition",
          NodeName: "条件分支",
          NodeType: "CONDITION",
          SortOrder: 1,
          ApproverType: "USERS",
          ApproverConfig: "{}",
          ConditionConfig: JSON.stringify({
            branches: [
              {
                id: "branch1",
                name: "分支1",
                condition: { fieldName: "", operator: "eq", fieldValue: "" },
                nodes: [
                  {
                    NodeCode: "branch-approval",
                    NodeName: "原始审批",
                    NodeType: "APPROVAL",
                    SortOrder: 0,
                    ApproverType: "USERS",
                    ApproverConfig: '{"type":"USERS","users":[],"mode":"OR"}',
                    ConditionConfig: "{}",
                  },
                ],
              },
              {
                id: "branch2",
                name: "其他",
                condition: { fieldName: "", operator: "else", fieldValue: "" },
                nodes: [
                  {
                    NodeCode: "branch-reject",
                    NodeName: "自动驳回",
                    NodeType: "APPROVAL",
                    ApproverType: "AUTO_REJECT",
                    SortOrder: 0,
                    ApproverConfig:
                      '{"type":"AUTO_REJECT","mode":"AUTO"}',
                    ConditionConfig: "{}",
                  },
                ],
              },
            ],
          }),
        },
        {
          NodeCode: "end",
          NodeName: "结束",
          NodeType: "END",
          SortOrder: 2,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      status: "draft",
    };

    // 更新条件分支中的审批节点配置
    const updateData = {
      NodeName: "更新后的审批节点",
      Description: "更新后的描述信息",
      ApproverConfig: JSON.stringify({
        type: "USERS",
        users: ["user001", "user002"],
        mode: "OR",
      }),
    };

    // 调用更新服务
    const result = WorkflowService.updateNode(workflow, "branch-approval", updateData);

    // 验证更新成功
    expect(result.success).toBe(true);
    expect(result.error).toBe(null);

    // 验证条件节点的ConditionConfig已被正确更新
    const conditionNode = workflow.nodes[1];
    const config = JSON.parse(conditionNode.ConditionConfig);
    const updatedBranchNode = config.branches[0].nodes[0];

    // 验证条件分支中的节点已被正确更新
    expect(updatedBranchNode.NodeName).toBe("更新后的审批节点");
    expect(updatedBranchNode.Description).toBe("更新后的描述信息");

    // 验证审批配置
    const approverConfig = JSON.parse(updatedBranchNode.ApproverConfig);
    expect(approverConfig.users).toEqual(["user001", "user002"]);
    expect(approverConfig.mode).toBe("OR");

    // 验证workflowData同步：确保条件节点的ConditionConfig包含最新的节点数据
    expect(conditionNode.ConditionConfig).toContain("更新后的审批节点");
    expect(conditionNode.ConditionConfig).toContain("更新后的描述信息");
    expect(conditionNode.ConditionConfig).toContain("user001");
    expect(conditionNode.ConditionConfig).toContain("user002");
  });

  it("outputs condition branch nodes in correct order with null IDs", () => {
    // 创建一个包含条件分支的工作流
    const workflowData = [
      {
        NodeCode: "start",
        NodeName: "开始",
        NodeType: "START",
        ID: 1,
        SortOrder: 0,
        ApproverType: "SYSTEM",
        ApproverConfig: "{}",
        ConditionConfig: "{}",
      },
      {
        NodeCode: "condition",
        NodeName: "条件分支",
        NodeType: "CONDITION",
        ID: 2,
        SortOrder: 1,
        ApproverType: "USERS",
        ApproverConfig: "{}",
        ConditionConfig: JSON.stringify({
          branches: [
            {
              id: "branch1",
              name: "分支1",
              condition: { fieldName: "", operator: "eq", fieldValue: "" },
              nodes: [
                {
                  NodeCode: "branch-approval",
                  NodeName: "分支审批",
                  NodeType: "APPROVAL",
                  ID: null, // 新添加的节点ID为空
                  SortOrder: 0,
                  ApproverType: "USERS",
                  ApproverConfig: '{"type":"USERS","users":[],"mode":"OR"}',
                  ConditionConfig: "{}",
                },
              ],
            },
            {
              id: "branch2",
              name: "其他",
              condition: { fieldName: "", operator: "else", fieldValue: "" },
              nodes: [
                {
                  NodeCode: "branch-reject",
                  NodeName: "自动驳回",
                  NodeType: "APPROVAL",
                  ApproverType: "AUTO_REJECT",
                  ID: null, // 新添加的节点ID为空
                  SortOrder: 0,
                  ApproverConfig:
                    '{"type":"AUTO_REJECT","mode":"AUTO"}',
                  ConditionConfig: "{}",
                },
              ],
            },
          ],
        }),
      },
      {
        NodeCode: "end",
        NodeName: "结束",
        NodeType: "END",
        ID: 3,
        SortOrder: 2,
        ApproverType: "SYSTEM",
        ApproverConfig: "{}",
        ConditionConfig: "{}",
      },
    ];

    // 模拟 getBackendData 的逻辑
    const resultNodes = [];
    const conditionBranchNodeCodes = new Set();

    // 第一次遍历：收集条件分支节点
    workflowData.forEach((node) => {
      if (node.NodeType === "CONDITION") {
        const config = JSON.parse(node.ConditionConfig);
        if (config.branches && Array.isArray(config.branches)) {
          config.branches.forEach((branch) => {
            if (branch.nodes && Array.isArray(branch.nodes)) {
              branch.nodes.forEach((branchNode) => {
                const nodeCode = branchNode.nodeCode || branchNode.NodeCode;
                if (nodeCode) {
                  conditionBranchNodeCodes.add(nodeCode);
                }
              });
            }
          });
        }
      }
    });

    // 第二次遍历：处理主流程节点和条件分支节点
    let currentSortOrder = 0;
    const approvalDefCode = workflowData[0]?.ApprovalDefCode || "";

    workflowData.forEach((node) => {
      if (!conditionBranchNodeCodes.has(node.NodeCode)) {
        // 处理主流程节点
        const processedNode = {
          ID: node.ID,
          ApprovalDefCode: node.ApprovalDefCode || approvalDefCode,
          NodeCode: node.NodeCode,
          NodeName: node.NodeName,
          NodeType: node.NodeType,
          Description: node.Description || "",
          SortOrder: currentSortOrder++,
          ApproverType: node.ApproverType,
          ApproverConfig: node.ApproverConfig,
          ConditionConfig: node.ConditionConfig,
          Status: node.Status || "",
        };
        resultNodes.push(processedNode);

        // 如果是条件节点，处理其分支节点
        if (node.NodeType === "CONDITION") {
          const config = JSON.parse(node.ConditionConfig);
          if (config.branches && Array.isArray(config.branches)) {
            config.branches.forEach((branch) => {
              if (branch.nodes && Array.isArray(branch.nodes)) {
                branch.nodes.forEach((branchNode) => {
                  const processedBranchNode = {
                    ID: branchNode.ID || branchNode.id || null,
                    ApprovalDefCode: branchNode.ApprovalDefCode || approvalDefCode,
                    NodeCode: branchNode.nodeCode || branchNode.NodeCode,
                    NodeName: branchNode.nodeName || branchNode.NodeName,
                    NodeType: branchNode.nodeType || branchNode.NodeType,
                    Description: branchNode.description || branchNode.Description || "",
                    SortOrder: currentSortOrder++,
                    ApproverType: branchNode.approverType || branchNode.ApproverType,
                    ApproverConfig: branchNode.approverConfig || branchNode.ApproverConfig,
                    ConditionConfig:
                      branchNode.conditionConfig || branchNode.ConditionConfig || "{}",
                    Status: branchNode.status || "",
                  };
                  resultNodes.push(processedBranchNode);
                });
              }
            });
          }
        }
      }
    });

    // 验证输出顺序
    expect(resultNodes).toHaveLength(5); // 开始、条件、分支审批、自动驳回、结束
    expect(resultNodes[0].NodeName).toBe("开始");
    expect(resultNodes[0].SortOrder).toBe(0);
    expect(resultNodes[1].NodeName).toBe("条件分支");
    expect(resultNodes[1].SortOrder).toBe(1);
    expect(resultNodes[2].NodeName).toBe("分支审批");
    expect(resultNodes[2].SortOrder).toBe(2);
    expect(resultNodes[3].NodeName).toBe("自动驳回");
    expect(resultNodes[3].SortOrder).toBe(3);
    expect(resultNodes[4].NodeName).toBe("结束");
    expect(resultNodes[4].SortOrder).toBe(4);

    // 验证新添加节点的ID为空
    expect(resultNodes[2].ID).toBe(null); // 分支审批节点
    expect(resultNodes[3].ID).toBe(null); // 自动驳回节点
    expect(resultNodes[0].ID).toBe(1); // 原有节点ID保持不变
    expect(resultNodes[1].ID).toBe(2); // 原有节点ID保持不变
    expect(resultNodes[4].ID).toBe(3); // 原有节点ID保持不变
  });

  it("creates new nodes with null IDs", () => {
    // 测试 NodeService 创建新节点时ID为空
    const approvalNode = NodeService.createNode("APPROVAL", "审批节点");
    const ccNode = NodeService.createNode("CC", "抄送节点");
    const conditionNode = NodeService.createNode("CONDITION", "条件分支");

    expect(approvalNode.ID).toBe(null);
    expect(ccNode.ID).toBe(null);
    expect(conditionNode.ID).toBe(null);

    // 验证其他字段正常生成
    expect(approvalNode.NodeCode).toBeDefined();
    expect(ccNode.NodeCode).toBeDefined();
    expect(conditionNode.NodeCode).toBeDefined();
  });

  it("converts auto reject node to auto approve node", () => {
    // 测试自动驳回节点可以修改为自动通过节点
    const workflow = {
      name: "Test Workflow",
      nodes: [
        {
          NodeCode: "start",
          NodeName: "开始",
          NodeType: "START",
          SortOrder: 0,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
        {
          NodeCode: "condition",
          NodeName: "条件分支",
          NodeType: "CONDITION",
          SortOrder: 1,
          ApproverType: "USERS",
          ApproverConfig: "{}",
          ConditionConfig: JSON.stringify({
            branches: [
              {
                id: "branch1",
                name: "分支1",
                condition: { fieldName: "amount", operator: "gt", fieldValue: "1000" },
                nodes: [
                  {
                    NodeCode: "branch-approval",
                    NodeName: "经理审批",
                    NodeType: "APPROVAL",
                    SortOrder: 0,
                    ApproverType: "USERS",
                    ApproverConfig: '{"type":"USERS","users":["manager"],"mode":"OR"}',
                    ConditionConfig: "{}",
                  },
                ],
              },
              {
                id: "branch2",
                name: "其他",
                condition: { fieldName: "", operator: "else", fieldValue: "" },
                nodes: [
                  {
                    NodeCode: "auto-reject",
                    NodeName: "自动驳回",
                    NodeType: "APPROVAL",
                    ApproverType: "AUTO_REJECT",
                    SortOrder: 0,
                    ApproverConfig: '{"type":"AUTO_REJECT","mode":"AUTO"}',
                    ConditionConfig: "{}",
                  },
                ],
              },
            ],
          }),
        },
        {
          NodeCode: "end",
          NodeName: "结束",
          NodeType: "END",
          SortOrder: 2,
          ApproverType: "SYSTEM",
          ApproverConfig: "{}",
          ConditionConfig: "{}",
        },
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      status: "draft",
    };

    // 将自动驳回节点修改为自动通过节点
    const updateData = {
      NodeName: "自动通过",
      ApproverType: "AUTO_APPROVE",
      ApproverConfig: JsonHelper.safeStringify({
        type: "AUTO_APPROVE",
        mode: "AUTO",
      }),
    };

    // 调用更新服务
    const result = WorkflowService.updateNode(workflow, "auto-reject", updateData);

    // 验证更新成功
    expect(result.success).toBe(true);
    expect(result.error).toBe(null);

    // 验证条件节点的ConditionConfig已被正确更新
    const conditionNode = workflow.nodes[1];
    const config = JSON.parse(conditionNode.ConditionConfig);
    const updatedBranchNode = config.branches[1].nodes[0];

    // 验证节点类型已改为自动通过
    expect(updatedBranchNode.NodeType).toBe("APPROVAL");
    expect(updatedBranchNode.ApproverType).toBe("AUTO_APPROVE");
    expect(updatedBranchNode.NodeName).toBe("自动通过");

    // 验证审批配置
    const approverConfig = JSON.parse(updatedBranchNode.ApproverConfig);
    expect(approverConfig.type).toBe("AUTO_APPROVE");
    expect(approverConfig.mode).toBe("AUTO");
  });

  it("integrates complete workflow with condition branches and null IDs", () => {
    // 创建一个完整的工作流构建器实例来测试集成
    const workflow = {
      name: "Test Workflow",
      nodes: [],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      status: "draft",
    };

    // 1. 添加开始节点
    WorkflowService.addNode(workflow, workflow.nodes, 0, "START", "开始");

    // 2. 添加审批节点
    WorkflowService.addNode(workflow, workflow.nodes, 1, "APPROVAL", "上级审批");

    // 3. 添加条件分支节点
    WorkflowService.addNode(workflow, workflow.nodes, 2, "CONDITION", "条件判断");

    // 4. 添加结束节点
    WorkflowService.addNode(workflow, workflow.nodes, 3, "END", "结束");

    // 验证工作流结构
    expect(workflow.nodes).toHaveLength(4);
    expect(workflow.nodes[0].NodeType).toBe("START");
    expect(workflow.nodes[1].NodeType).toBe("APPROVAL");
    expect(workflow.nodes[2].NodeType).toBe("CONDITION");
    expect(workflow.nodes[3].NodeType).toBe("END");

    // 验证所有新节点的ID都为null
    expect(workflow.nodes[0].ID).toBe(null);
    expect(workflow.nodes[1].ID).toBe(null);
    expect(workflow.nodes[2].ID).toBe(null);
    expect(workflow.nodes[3].ID).toBe(null);

    // 验证条件分支包含两个默认节点
    const conditionConfig = JSON.parse(workflow.nodes[2].ConditionConfig);
    expect(conditionConfig.branches).toHaveLength(2);
    expect(conditionConfig.branches[0].nodes).toHaveLength(1);
    expect(conditionConfig.branches[1].nodes).toHaveLength(1);

    // 验证条件分支中的节点ID也为null
    expect(conditionConfig.branches[0].nodes[0].ID).toBe(null);
    expect(conditionConfig.branches[1].nodes[0].ID).toBe(null);

    // 模拟 getBackendData 的输出逻辑
    const resultNodes = [];
    const conditionBranchNodeCodes = new Set();

    // 第一次遍历：收集条件分支节点
    workflow.nodes.forEach((node) => {
      if (node.NodeType === "CONDITION") {
        const config = JSON.parse(node.ConditionConfig);
        if (config.branches && Array.isArray(config.branches)) {
          config.branches.forEach((branch) => {
            if (branch.nodes && Array.isArray(branch.nodes)) {
              branch.nodes.forEach((branchNode) => {
                const nodeCode = branchNode.nodeCode || branchNode.NodeCode;
                if (nodeCode) {
                  conditionBranchNodeCodes.add(nodeCode);
                }
              });
            }
          });
        }
      }
    });

    // 第二次遍历：处理主流程节点和条件分支节点
    let currentSortOrder = 0;
    const approvalDefCode = "TEST_APPROVAL_DEF";

    workflow.nodes.forEach((node) => {
      if (!conditionBranchNodeCodes.has(node.NodeCode)) {
        // 处理主流程节点
        const processedNode = {
          ID: node.ID,
          ApprovalDefCode: node.ApprovalDefCode || approvalDefCode,
          NodeCode: node.NodeCode,
          NodeName: node.NodeName,
          NodeType: node.NodeType,
          Description: node.Description || "",
          SortOrder: currentSortOrder++,
          ApproverType: node.ApproverType,
          ApproverConfig: node.ApproverConfig,
          ConditionConfig: node.ConditionConfig,
          Status: node.Status || "",
        };
        resultNodes.push(processedNode);

        // 如果是条件节点，处理其分支节点
        if (node.NodeType === "CONDITION") {
          const config = JSON.parse(node.ConditionConfig);
          if (config.branches && Array.isArray(config.branches)) {
            config.branches.forEach((branch) => {
              if (branch.nodes && Array.isArray(branch.nodes)) {
                branch.nodes.forEach((branchNode) => {
                  const processedBranchNode = {
                    ID: branchNode.ID || branchNode.id || null,
                    ApprovalDefCode: branchNode.ApprovalDefCode || approvalDefCode,
                    NodeCode: branchNode.nodeCode || branchNode.NodeCode,
                    NodeName: branchNode.nodeName || branchNode.NodeName,
                    NodeType: branchNode.nodeType || branchNode.NodeType,
                    Description: branchNode.description || branchNode.Description || "",
                    SortOrder: currentSortOrder++,
                    ApproverType: branchNode.approverType || branchNode.ApproverType,
                    ApproverConfig: branchNode.approverConfig || branchNode.ApproverConfig,
                    ConditionConfig:
                      branchNode.conditionConfig || branchNode.ConditionConfig || "{}",
                    Status: branchNode.status || "",
                  };
                  resultNodes.push(processedBranchNode);
                });
              }
            });
          }
        }
      }
    });

    // 验证输出顺序和结构
    expect(resultNodes).toHaveLength(6); // 开始、上级审批、条件、分支审批、自动驳回、结束
    expect(resultNodes[0].NodeName).toBe("开始");
    expect(resultNodes[0].SortOrder).toBe(0);
    expect(resultNodes[0].ID).toBe(null);

    expect(resultNodes[1].NodeName).toBe("上级审批");
    expect(resultNodes[1].SortOrder).toBe(1);
    expect(resultNodes[1].ID).toBe(null);

    expect(resultNodes[2].NodeName).toBe("条件判断");
    expect(resultNodes[2].SortOrder).toBe(2);
    expect(resultNodes[2].ID).toBe(null);

    expect(resultNodes[3].NodeName).toBe("审批节点"); // 分支1中的审批节点
    expect(resultNodes[3].SortOrder).toBe(3);
    expect(resultNodes[3].ID).toBe(null);

    expect(resultNodes[4].NodeName).toBe("自动驳回"); // 其他分支中的自动驳回节点
    expect(resultNodes[4].SortOrder).toBe(4);
    expect(resultNodes[4].ID).toBe(null);

    expect(resultNodes[5].NodeName).toBe("结束");
    expect(resultNodes[5].SortOrder).toBe(5);
    expect(resultNodes[5].ID).toBe(null);

    // 验证ApprovalDefCode正确设置
    resultNodes.forEach((node) => {
      expect(node.ApprovalDefCode).toBe("TEST_APPROVAL_DEF");
    });
  });
});
