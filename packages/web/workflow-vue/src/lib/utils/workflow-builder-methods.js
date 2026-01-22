import { WorkflowUtils } from './workflow-utils.js';
import { JsonHelper } from './json-helper.js';
import { v4 as uuidv4 } from 'uuid';

// 从 WorkflowBuilder.vue 中提取的方法，避免代码重复

// 默认工作流数据
export const defaultWorkflowData = [
  {
    NodeName: '提交',
    NodeType: 'START',
    Description: '',
    SortOrder: 0,
    ApproverType: 'USERS',
    ApproverConfig: '{}',
    ConditionConfig: '{}',
    Status: 'Normal',
  },
  {
    NodeName: '上级审批',
    NodeType: 'APPROVAL',
    Description: '',
    SortOrder: 1,
    ApproverType: 'USERS',
    ApproverConfig: '{"type":"USERS","users":[],"mode":"OR"}',
    ConditionConfig: '{}',
    Status: 'Normal',
  },
  {
    NodeName: '结束',
    NodeType: 'END',
    Description: '',
    ApproverType: 'USERS',
    ApproverConfig: '{}',
    ConditionConfig: '{}',
    Status: 'Normal',
  },
];

// 从后端数据加载工作流
export const loadFromBackendData = (workflowData, backendData) => {
  try {
    // 如果没有输入数据或数据无效，则使用默认数据
    const dataToProcess = Array.isArray(backendData) ? backendData : defaultWorkflowData;

    // 处理数据，确保每个节点都有必要的字段
    const processedData = dataToProcess.map(node => {
      return {
        // 确保必要字段存在
        ID: node.ID,
        NodeCode: node.NodeCode || uuidv4(),
        NodeName: node.NodeName || '未命名节点',
        NodeType: node.NodeType || 'APPROVAL',
        Description: node.Description || '',
        SortOrder: node.SortOrder || 0,
        ApproverType: node.ApproverType || 'USERS',
        ApproverConfig: node.ApproverConfig || '{}',
        ConditionConfig: node.ConditionConfig || '{}',
        ApprovalDefCode: node.ApprovalDefCode || '',
        Status: node.Status || 'Normal',
        // 保留原始数据的其他字段
        ...node,
      };
    });

    // 使用工具函数处理数据
    const finalData = WorkflowUtils.processConditionConfig(processedData);

    // 更新响应式数据
    workflowData.splice(0, workflowData.length, ...finalData);

    // 保存原始backendData的ID映射，供getBackendData使用
    workflowData._originalNodeCodeToIdMap = new Map();
    dataToProcess.forEach(node => {
      if (node.NodeCode && node.ID) {
        workflowData._originalNodeCodeToIdMap.set(node.NodeCode, node.ID);
      }
    });
  } catch (error) {
    // 发生错误时使用默认数据
    const fallbackData = WorkflowUtils.processConditionConfig(defaultWorkflowData);
    workflowData.splice(0, workflowData.length, ...fallbackData);
  }
};

// 获取处理后端数据
export const getBackendData = workflowData => {
  if (!Array.isArray(workflowData)) {
    return [];
  }

  const resultNodes = [];
  const conditionBranchNodeCodes = new Set();

  // 使用保存的原始数据映射
  const nodeCodeToIdMap = workflowData._originalNodeCodeToIdMap || new Map();

  // 第一次遍历：收集条件分支节点
  workflowData.forEach(node => {
    if (node.NodeType === 'CONDITION') {
      const config = JsonHelper.safeParse(node.ConditionConfig);
      if (config.branches && Array.isArray(config.branches)) {
        config.branches.forEach(branch => {
          if (branch.nodes && Array.isArray(branch.nodes)) {
            branch.nodes.forEach(branchNode => {
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
  const approvalDefCode = workflowData[0]?.ApprovalDefCode || '';

  workflowData.forEach(node => {
    if (!conditionBranchNodeCodes.has(node.NodeCode)) {
      // 处理主流程节点
      let finalNodeType = node.NodeType;
      const finalApproverType = node.ApproverType;
      if (finalApproverType === 'AUTO_APPROVE' || finalApproverType === 'AUTO_REJECT') {
        finalNodeType = 'APPROVAL';
      }

      let finalConditionConfig = node.ConditionConfig;
      if (node.NodeType === 'CONDITION') {
        const config = JsonHelper.safeParse(node.ConditionConfig);
        if (config.branches && Array.isArray(config.branches)) {
          config.branches.forEach(branch => {
            if (branch.nodes && Array.isArray(branch.nodes)) {
              branch.nodes.forEach(branchNode => {
                const approverType = branchNode.approverType || branchNode.ApproverType;
                if (approverType === 'AUTO_APPROVE' || approverType === 'AUTO_REJECT') {
                  branchNode.nodeType = 'APPROVAL';
                  branchNode.NodeType = 'APPROVAL';
                }
              });
            }
          });
          finalConditionConfig = JSON.stringify(config);
        }
      }

      const processedNode = {
        ID: node.ID,
        ApprovalDefCode: node.ApprovalDefCode || approvalDefCode,
        NodeCode: node.NodeCode,
        NodeName: node.NodeName,
        NodeType: finalNodeType,
        Description: node.Description || '',
        SortOrder: currentSortOrder++,
        ApproverType: finalApproverType,
        ApproverConfig: node.ApproverConfig,
        ConditionConfig: finalConditionConfig,
        Status: node.Status || 'Normal',
      };
      resultNodes.push(processedNode);

      // 如果是条件节点，处理其分支节点
      if (node.NodeType === 'CONDITION') {
        const config = JsonHelper.safeParse(node.ConditionConfig);
        if (config.branches && Array.isArray(config.branches)) {
          config.branches.forEach(branch => {
            if (branch.nodes && Array.isArray(branch.nodes)) {
              branch.nodes.forEach(branchNode => {
                const branchNodeCode = branchNode.nodeCode || branchNode.NodeCode;

                // ID处理分为两种情况：
                // 1. 如果条件分支中的节点是新添加的，ID为空
                // 2. 如果条件分支的数据已经存在，根据 nodeCode 查找已存在的ID
                let branchNodeId = null;
                if (nodeCodeToIdMap.has(branchNodeCode)) {
                  // 优先使用映射中的ID（来自主流程节点）
                  branchNodeId = nodeCodeToIdMap.get(branchNodeCode);
                } else if (branchNode.ID || branchNode.id) {
                  // 如果分支节点本身有ID，使用分支节点的ID
                  branchNodeId = branchNode.ID || branchNode.id;
                }
                // 否则保持为null（新添加的节点）

                const finalBranchApproverType = branchNode.approverType || branchNode.ApproverType;
                let finalBranchNodeType = branchNode.nodeType || branchNode.NodeType;
                if (
                  finalBranchApproverType === 'AUTO_APPROVE' ||
                  finalBranchApproverType === 'AUTO_REJECT'
                ) {
                  finalBranchNodeType = 'APPROVAL';
                }

                const processedBranchNode = {
                  ID: branchNodeId,
                  ApprovalDefCode: branchNode.ApprovalDefCode || approvalDefCode,
                  NodeCode: branchNodeCode,
                  NodeName: branchNode.nodeName || branchNode.NodeName,
                  NodeType: finalBranchNodeType,
                  Description: branchNode.description || branchNode.Description || '',
                  SortOrder: currentSortOrder++,
                  ApproverType: finalBranchApproverType,
                  ApproverConfig: branchNode.approverConfig || branchNode.ApproverConfig,
                  // 条件分支节点继承并保存所在分支的条件配置
                  ConditionConfig: branch.condition
                    ? JSON.stringify(branch.condition)
                    : branchNode.conditionConfig || branchNode.ConditionConfig || '{}',
                  Status: branchNode.status || 'Normal',
                };
                resultNodes.push(processedBranchNode);
              });
            }
          });
        }
      }
    }
  });

  // 按 SortOrder 排序
  resultNodes.sort((a, b) => a.SortOrder - b.SortOrder);

  return resultNodes;
};
