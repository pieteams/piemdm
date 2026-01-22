import { describe, it, expect, beforeEach } from "vitest";
import { reactive } from "vue";
import { WorkflowService } from "../src/lib/services/workflow-service.js";
import { WorkflowUtils } from "../src/lib/utils/workflow-utils.js";
import { JsonHelper } from "../src/lib/utils/json-helper.js";
import {
  defaultWorkflowData,
  loadFromBackendData,
  getBackendData,
} from "../src/lib/utils/workflow-builder-methods.js";

describe("Approval Definition Integration with WorkflowBuilder", () => {
  let mockWorkflowData;

  // 保留原始的测试数据，不删除
  const workflowData = [
    {
      ID: 178,
      ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
      NodeCode: "566fcc47-e977-40fb-b202-b91ed2caf37c",
      NodeName: "提交",
      NodeType: "START",
      Description: "",
      SortOrder: 0,
      ApproverType: "SYSTEM",
      ApproverConfig: "{}",
      ConditionConfig: "{}",
    },
    {
      ID: 179,
      ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
      NodeCode: "ec7e63bc-b1b0-475e-8961-1facfd88db2c",
      NodeName: "上级审批",
      NodeType: "APPROVAL",
      Description: "",
      SortOrder: 1,
      ApproverType: "USERS",
      ApproverConfig: '{"type":"USERS","users":["jasen"],"mode":"OR"}',
      ConditionConfig: "{}",
    },
    {
      ID: 180,
      ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
      NodeCode: "a88b9b61-490e-4507-afd0-89f18a6f121e",
      NodeName: "条件分支",
      NodeType: "CONDITION",
      Description: "",
      SortOrder: 2,
      ApproverType: "USERS",
      ApproverConfig: "{}",
      ConditionConfig:
        '{"branches":[{"name":"条件分支 1","condition":{"fieldName":"price","operator":"gte","fieldValue":"3000"},"nodes":[{"nodeCode":"fa8b52fe-659d-43fb-812d-a9a6b0f3f6dc","nodeName":"财务审批","nodeType":"APPROVAL","sortOrder":1,"approvalDefCode":"TEST_APPROVAL_1749042000727964000","approverType":"USERS","approverConfig":"{\\\"type\\\":\\\"USERS\\\",\\\"users\\\":[\\\"jasen\\\"],\\\"mode\\\":\\\"OR\\\"}","approvalMode":"OR","conditionExpr":"","conditionConfig":"{}","timeoutHours":72,"autoRemind":true,"remindInterval":24,"nextNodes":"","rejectToNode":"","rejectToStart":false,"notifyConfig":"{}","enableEmail":true,"enableSMS":false,"formFields":"{}","readOnlyFields":"","requiredFields":"","visibleToAll":false,"allowedRoles":"","allowedDepts":"","customConfig":"{}","tags":"","status":"Active","version":1}]},{"name":"其他情况","condition":{"fieldName":"","operator":"eq","fieldValue":""},"nodes":[{"nodeCode":"73e194ac-3de1-4545-91d7-4f727acca413","nodeName":"自动驳回","nodeType":"APPROVAL","sortOrder":1,"approvalDefCode":"TEST_APPROVAL_1749042000727964000","approverType":"AUTO_REJECT","approverConfig":"{\\\"type\\\":\\\"USERS\\\",\\\"users\\\":[],\\\"mode\\\":\\\"OR\\\"}"}]}]}',
    },
    {
      ID: 181,
      ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
      NodeCode: "fa8b52fe-659d-43fb-812d-a9a6b0f3f6dc",
      NodeName: "财务审批",
      NodeType: "APPROVAL",
      Description: "",
      SortOrder: 3,
      ApproverType: "USERS",
      ApproverConfig: '{"type":"USERS","users":["jasen"],"mode":"OR"}',
      ConditionConfig: "{}",
    },
    {
      ID: 182,
      ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
      NodeCode: "73e194ac-3de1-4545-91d7-4f727acca413",
      NodeName: "自动驳回",
      NodeType: "APPROVAL",
      Description: "",
      SortOrder: 4,
      ApproverType: "AUTO_REJECT",
      ApproverConfig: '{"type":"USERS","users":[],"mode":"OR"}',
      ConditionConfig: "{}",
    },
    {
      ID: 183,
      ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
      NodeCode: "565a5219-23bc-4d74-af86-42ba2ff79d77",
      NodeName: "抄送法务",
      NodeType: "CC",
      Description: "",
      SortOrder: 5,
      ApproverType: "USERS",
      ApproverConfig:
        '{"type":"CC","users":["jasen","jasen888"],"mode":"CC","ccTiming":"after_approval"}',
      ConditionConfig: "{}",
    },
    {
      ID: 184,
      ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
      NodeCode: "d96b927a-9a05-41ce-bb22-f9f4d0aa51b5",
      NodeName: "结束",
      NodeType: "END",
      Description: "",
      SortOrder: 6,
      ApproverType: "SYSTEM",
      ApproverConfig: "{}",
      ConditionConfig: "{}",
    },
  ];

  beforeEach(() => {
    // 初始化响应式数据
    mockWorkflowData = reactive(WorkflowUtils.processConditionConfig(defaultWorkflowData));
  });

  it("loads real workflow data using loadFromBackendData", () => {
    // 调用从 WorkflowBuilder.vue 提取的 loadFromBackendData 方法
    loadFromBackendData(mockWorkflowData, workflowData);

    // 获取加载后的数据
    const loadedData = Array.from(mockWorkflowData);

    // 验证数据结构 - 应该处理条件分支，将分支节点包含在条件节点中
    expect(loadedData).toHaveLength(5); // 提交、上级审批、条件、抄送法务、结束（条件分支中的节点被包含在条件节点中）

    // 验证主流程节点
    expect(loadedData[0].NodeName).toBe("提交");
    expect(loadedData[0].NodeType).toBe("START");
    expect(loadedData[0].ID).toBe(178);

    expect(loadedData[1].NodeName).toBe("上级审批");
    expect(loadedData[1].NodeType).toBe("APPROVAL");
    expect(loadedData[1].ID).toBe(179);

    expect(loadedData[2].NodeName).toBe("条件分支");
    expect(loadedData[2].NodeType).toBe("CONDITION");
    expect(loadedData[2].ID).toBe(180);

    // 验证条件配置
    const conditionConfig = JsonHelper.safeParse(loadedData[2].ConditionConfig);
    expect(conditionConfig.branches).toHaveLength(2);
    expect(conditionConfig.branches[0].name).toBe("条件分支 1");
    expect(conditionConfig.branches[0].nodes).toHaveLength(1);
    expect(conditionConfig.branches[0].nodes[0].nodeName).toBe("财务审批");
    expect(conditionConfig.branches[1].name).toBe("其他情况");
    expect(conditionConfig.branches[1].nodes).toHaveLength(1);
    expect(conditionConfig.branches[1].nodes[0].nodeName).toBe("自动驳回");

    expect(loadedData[3].NodeName).toBe("抄送法务");
    expect(loadedData[3].NodeType).toBe("CC");
    expect(loadedData[3].ID).toBe(183);

    expect(loadedData[4].NodeName).toBe("结束");
    expect(loadedData[4].NodeType).toBe("END");
    expect(loadedData[4].ID).toBe(184);
  });

  it("outputs backend data using getBackendData", () => {
    // 先加载数据
    loadFromBackendData(mockWorkflowData, workflowData);

    // 调用从 WorkflowBuilder.vue 提取的 getBackendData 方法
    const backendData = getBackendData(mockWorkflowData);

    // 验证输出顺序和结构
    expect(backendData).toHaveLength(7); // 提交、上级审批、条件、财务审批、自动驳回、抄送法务、结束

    expect(backendData[0].NodeName).toBe("提交");
    expect(backendData[0].SortOrder).toBe(0);
    expect(backendData[0].ID).toBe(178);

    expect(backendData[1].NodeName).toBe("上级审批");
    expect(backendData[1].SortOrder).toBe(1);
    expect(backendData[1].ID).toBe(179);

    expect(backendData[2].NodeName).toBe("条件分支");
    expect(backendData[2].SortOrder).toBe(2);
    expect(backendData[2].ID).toBe(180);

    expect(backendData[3].NodeName).toBe("财务审批");
    expect(backendData[3].SortOrder).toBe(3);
    expect(backendData[3].ID).toBe(181); // 条件分支中的节点ID从主流程数据中获取

    expect(backendData[4].NodeName).toBe("自动驳回");
    expect(backendData[4].SortOrder).toBe(4);
    expect(backendData[4].ID).toBe(182); // 条件分支中的节点ID从主流程数据中获取

    expect(backendData[5].NodeName).toBe("抄送法务");
    expect(backendData[5].SortOrder).toBe(5);
    expect(backendData[5].ID).toBe(183);

    expect(backendData[6].NodeName).toBe("结束");
    expect(backendData[6].SortOrder).toBe(6);
    expect(backendData[6].ID).toBe(184);

    // 验证审批配置
    const approvalConfig1 = JsonHelper.safeParse(backendData[1].ApproverConfig);
    expect(approvalConfig1.users).toEqual(["jasen"]);
    expect(approvalConfig1.mode).toBe("OR");

    const ccConfig = JsonHelper.safeParse(backendData[5].ApproverConfig);
    expect(ccConfig.type).toBe("CC");
    expect(ccConfig.users).toEqual(["jasen", "jasen888"]);
    expect(ccConfig.ccTiming).toBe("after_approval");

    const autoRejectConfig = JsonHelper.safeParse(backendData[4].ApproverConfig);
    expect(autoRejectConfig.type).toBe("USERS"); // 原始数据中是USERS，但节点类型是AUTO_REJECT
    expect(autoRejectConfig.users).toEqual([]);
  });

  it("updates node configuration through WorkflowBuilder", () => {
    // 加载数据
    loadFromBackendData(mockWorkflowData, workflowData);

    // 验证初始状态
    let loadedData = Array.from(mockWorkflowData);
    expect(loadedData[1].NodeName).toBe("上级审批");

    // 更新节点配置
    const updateData = {
      NodeName: "部门经理审批",
      ApproverConfig: '{"type":"USERS","users":["manager","hr"],"mode":"AND"}',
    };

    // 使用 WorkflowService 更新节点
    const workflow = {
      nodes: mockWorkflowData,
    };

    const result = WorkflowService.updateNode(workflow, loadedData[1].NodeCode, updateData);
    expect(result.success).toBe(true);

    // 验证更新后的数据
    loadedData = Array.from(mockWorkflowData);
    expect(loadedData[1].NodeName).toBe("部门经理审批");

    const updatedConfig = JsonHelper.safeParse(loadedData[1].ApproverConfig);
    expect(updatedConfig.users).toEqual(["manager", "hr"]);
    expect(updatedConfig.mode).toBe("AND");
  });

  it("preserves original IDs from workflowData", () => {
    // 加载数据
    loadFromBackendData(mockWorkflowData, workflowData);

    // 验证原始ID被保留
    const loadedData = Array.from(mockWorkflowData);
    expect(loadedData[0].ID).toBe(178);
    expect(loadedData[1].ID).toBe(179);
    expect(loadedData[2].ID).toBe(180);
    expect(loadedData[3].ID).toBe(183);
    expect(loadedData[4].ID).toBe(184);
  });

  it("handles condition branch nodes correctly in backend output", () => {
    // 加载数据
    loadFromBackendData(mockWorkflowData, workflowData);

    // 获取后端数据
    const backendData = getBackendData(mockWorkflowData);

    // 验证条件分支节点在输出中的处理
    const conditionNode = backendData.find((node) => node.NodeType === "CONDITION");
    expect(conditionNode).toBeDefined();
    expect(conditionNode.NodeName).toBe("条件分支");
    expect(conditionNode.ID).toBe(180);

    // 验证条件分支中的节点被正确输出
    const branchNodes = backendData.filter(
      (node) =>
        node.NodeCode === "fa8b52fe-659d-43fb-812d-a9a6b0f3f6dc" ||
        node.NodeCode === "73e194ac-3de1-4545-91d7-4f727acca413",
    );

    expect(branchNodes).toHaveLength(2);
    expect(branchNodes[0].NodeName).toBe("财务审批");
    expect(branchNodes[1].NodeName).toBe("自动驳回");

    // 验证分支节点的ID从主流程数据中获取
    expect(branchNodes[0].ID).toBe(181);
    expect(branchNodes[1].ID).toBe(182);
  });

  it("validates complex condition configuration in real data", () => {
    // 加载数据
    loadFromBackendData(mockWorkflowData, workflowData);

    // 验证条件配置的复杂性
    const loadedData = Array.from(mockWorkflowData);
    const conditionNode = loadedData.find((node) => node.NodeType === "CONDITION");

    expect(conditionNode).toBeDefined();

    const conditionConfig = JsonHelper.safeParse(conditionNode.ConditionConfig);
    expect(conditionConfig.branches).toHaveLength(2);

    // 验证第一个条件分支
    const branch1 = conditionConfig.branches[0];
    expect(branch1.name).toBe("条件分支 1");
    expect(branch1.condition.fieldName).toBe("price");
    expect(branch1.condition.operator).toBe("gte");
    expect(branch1.condition.fieldValue).toBe("3000");
    expect(branch1.nodes).toHaveLength(1);
    expect(branch1.nodes[0].nodeName).toBe("财务审批");

    // 验证第二个条件分支
    const branch2 = conditionConfig.branches[1];
    expect(branch2.name).toBe("其他情况");
    expect(branch2.nodes).toHaveLength(1);
    expect(branch2.nodes[0].nodeName).toBe("自动驳回");
  });

  it("handles CC node configuration correctly in real data", () => {
    // 加载数据
    loadFromBackendData(mockWorkflowData, workflowData);

    // 验证抄送节点配置
    const loadedData = Array.from(mockWorkflowData);
    const ccNode = loadedData.find((node) => node.NodeType === "CC");

    expect(ccNode).toBeDefined();
    expect(ccNode.NodeName).toBe("抄送法务");

    const ccConfig = JsonHelper.safeParse(ccNode.ApproverConfig);
    expect(ccConfig.type).toBe("CC");
    expect(ccConfig.users).toEqual(["jasen", "jasen888"]);
    expect(ccConfig.ccTiming).toBe("after_approval");
  });

  it("maintains data integrity through load and output cycle", () => {
    // 加载数据
    loadFromBackendData(mockWorkflowData, workflowData);

    // 获取加载后的数据
    const loadedData = Array.from(mockWorkflowData);

    // 获取输出数据
    const outputData = getBackendData(mockWorkflowData);

    // 验证数据完整性
    expect(outputData.length).toBe(7);

    // 验证所有必要字段都存在
    outputData.forEach((node) => {
      expect(node).toHaveProperty("NodeCode");
      expect(node).toHaveProperty("NodeName");
      expect(node).toHaveProperty("NodeType");
      expect(node).toHaveProperty("SortOrder");
      expect(node).toHaveProperty("ApproverType");
      expect(node).toHaveProperty("ApproverConfig");
      expect(node).toHaveProperty("ConditionConfig");
    });

    // 验证排序正确
    for (let i = 0; i < outputData.length - 1; i++) {
      expect(outputData[i].SortOrder).toBeLessThanOrEqual(outputData[i + 1].SortOrder);
    }

    // 验证JSON配置可以正确解析
    outputData.forEach((node) => {
      expect(() => {
        JsonHelper.safeParse(node.ApproverConfig);
        JsonHelper.safeParse(node.ConditionConfig);
      }).not.toThrow();
    });
  });

  it("handles new condition branch nodes without existing IDs", () => {
    // 创建一个包含新节点的条件分支数据
    const workflowDataWithNewNodes = [
      {
        ID: 178,
        ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
        NodeCode: "566fcc47-e977-40fb-b202-b91ed2caf37c",
        NodeName: "提交",
        NodeType: "START",
        Description: "",
        SortOrder: 0,
        ApproverType: "SYSTEM",
        ApproverConfig: "{}",
        ConditionConfig: "{}",
      },
      {
        ID: 179,
        ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
        NodeCode: "ec7e63bc-b1b0-475e-8961-1facfd88db2c",
        NodeName: "上级审批",
        NodeType: "APPROVAL",
        Description: "",
        SortOrder: 1,
        ApproverType: "USERS",
        ApproverConfig: '{"type":"USERS","users":["jasen"],"mode":"OR"}',
        ConditionConfig: "{}",
      },
      {
        ID: 180,
        ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
        NodeCode: "a88b9b61-490e-4507-afd0-89f18a6f121e",
        NodeName: "条件分支",
        NodeType: "CONDITION",
        Description: "",
        SortOrder: 2,
        ApproverType: "USERS",
        ApproverConfig: "{}",
        ConditionConfig:
          '{"branches":[{"name":"条件分支 1","condition":{"fieldName":"price","operator":"gte","fieldValue":"3000"},"nodes":[{"nodeCode":"new-node-1","nodeName":"新财务审批","nodeType":"APPROVAL","sortOrder":1,"approvalDefCode":"TEST_APPROVAL_1749042000727964000","approverType":"USERS","approverConfig":"{\\\"type\\\":\\\"USERS\\\",\\\"users\\\":[\\\"jasen\\\"],\\\"mode\\\":\\\"OR\\\"}"}]},{"name":"其他情况","condition":{"fieldName":"","operator":"eq","fieldValue":""},"nodes":[{"nodeCode":"new-node-2","nodeName":"新自动驳回","nodeType":"APPROVAL","sortOrder":1,"approvalDefCode":"TEST_APPROVAL_1749042000727964000","approverType":"AUTO_REJECT","approverConfig":"{\\\"type\\\":\\\"USERS\\\",\\\"users\\\":[],\\\"mode\\\":\\\"OR\\\"}"}]}]}',
      },
      {
        ID: 183,
        ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
        NodeCode: "565a5219-23bc-4d74-af86-42ba2ff79d77",
        NodeName: "抄送法务",
        NodeType: "CC",
        Description: "",
        SortOrder: 5,
        ApproverType: "USERS",
        ApproverConfig:
          '{"type":"CC","users":["jasen","jasen888"],"mode":"CC","ccTiming":"after_approval"}',
        ConditionConfig: "{}",
      },
      {
        ID: 184,
        ApprovalDefCode: "TEST_APPROVAL_1749042000727964000",
        NodeCode: "d96b927a-9a05-41ce-bb22-f9f4d0aa51b5",
        NodeName: "结束",
        NodeType: "END",
        Description: "",
        SortOrder: 6,
        ApproverType: "SYSTEM",
        ApproverConfig: "{}",
        ConditionConfig: "{}",
      },
    ];

    // 加载包含新节点的数据
    loadFromBackendData(mockWorkflowData, workflowDataWithNewNodes);

    // 获取后端数据
    const backendData = getBackendData(mockWorkflowData);

    // 验证新节点的ID为null（因为没有对应的主流程数据）
    const newFinanceNode = backendData.find((node) => node.NodeCode === "new-node-1");
    expect(newFinanceNode).toBeDefined();
    expect(newFinanceNode.NodeName).toBe("新财务审批");
    expect(newFinanceNode.ID).toBe(null); // 新节点ID为null

    const newRejectNode = backendData.find((node) => node.NodeCode === "new-node-2");
    expect(newRejectNode).toBeDefined();
    expect(newRejectNode.NodeName).toBe("新自动驳回");
    expect(newRejectNode.ID).toBe(null); // 新节点ID为null
  });
});