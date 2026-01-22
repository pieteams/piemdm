/**
 * 工作流执行引擎
 * 负责工作流的执行、状态管理和流程控制
 */
import { NodeHelper } from '../utils/node-helper.js';
import { JsonHelper } from '../utils/json-helper.js';

export class WorkflowEngine {
  constructor(workflow) {
    this.workflow = workflow;
    this.executionState = {
      currentNodeCode: null,
      completedNodes: new Set(),
      failedNodes: new Set(),
      variables: {}, // 流程变量
      history: [], // 执行历史
      status: 'ready', // ready, running, completed, failed
    };
    this.startTime = null;
    this.endTime = null;
  }

  /**
   * 开始执行工作流
   * @param {Object} startData - 启动数据
   * @returns {Object} 执行结果
   */
  async start(startData = {}) {
    if (this.executionState.status !== 'ready') {
      throw new Error('工作流状态不正确，无法启动');
    }

    this.executionState.status = 'running';
    this.startTime = new Date();
    this.executionState.variables = { ...startData };

    try {
      // 找到开始节点
      const startNode = this.workflow.nodes.find(node => node.NodeType === 'START');
      if (!startNode) {
        throw new Error('工作流缺少开始节点');
      }

      // 记录执行历史
      this.addHistory('START', startNode, '流程开始');

      // 开始执行主流程
      const result = await this.executeFlow(this.workflow.nodes, startNode);

      this.endTime = new Date();
      this.executionState.status = result.success ? 'completed' : 'failed';

      return {
        success: result.success,
        message: result.message,
        executionTime: this.endTime - this.startTime,
        history: this.executionState.history,
        variables: this.executionState.variables,
      };
    } catch (error) {
      this.executionState.status = 'failed';
      this.endTime = new Date();

      return {
        success: false,
        message: error.message,
        executionTime: this.endTime - this.startTime,
        history: this.executionState.history,
        variables: this.executionState.variables,
      };
    }
  }

  /**
   * 执行流程
   * @param {Array} nodes - 节点列表
   * @param {Object} startNode - 开始节点
   * @returns {Object} 执行结果
   */
  async executeFlow(nodes, startNode) {
    let currentIndex = nodes.findIndex(node => node.NodeCode === startNode.NodeCode);
    let currentNode = startNode;

    while (currentNode && currentIndex < nodes.length) {
      this.executionState.currentNodeCode = currentNode.NodeCode;

      // 执行当前节点
      const nodeResult = await this.executeNode(currentNode);

      if (!nodeResult.success) {
        return {
          success: false,
          message: `节点 ${currentNode.NodeName} 执行失败: ${nodeResult.message}`,
          failedNode: currentNode,
        };
      }

      // 如果是结束节点，流程完成
      if (currentNode.NodeType === 'END') {
        this.addHistory('END', currentNode, '流程结束');
        return {
          success: true,
          message: '工作流执行完成',
        };
      }

      // 查找下一个节点
      const nextNode = this.findNextNode(nodes, currentIndex, currentNode);

      if (!nextNode) {
        return {
          success: false,
          message: `节点 ${currentNode.NodeName} 后续找不到有效节点`,
        };
      }

      currentNode = nextNode;
      currentIndex = nodes.findIndex(node => node.NodeCode === nextNode.NodeCode);
    }

    return {
      success: false,
      message: '流程异常结束',
    };
  }

  /**
   * 执行单个节点
   * @param {Object} node - 节点对象
   * @returns {Object} 执行结果
   */
  async executeNode(node) {
    const startTime = new Date();

    try {
      let result = { success: false, message: '' };

      switch (node.NodeType) {
        case 'START':
          result = await this.executeStartNode(node);
          break;
        case 'APPROVAL':
          result = await this.executeApprovalNode(node);
          break;
        case 'CONDITION':
          result = await this.executeConditionNode(node);
          break;
        case 'CC':
          result = await this.executeCCNode(node);
          break;
        case 'END':
          result = await this.executeEndNode(node);
          break;
        default:
          result = { success: false, message: `未知节点类型: ${node.NodeType}` };
      }

      const executionTime = new Date() - startTime;

      // 记录执行结果
      this.executionState.completedNodes.add(node.NodeCode);
      this.addHistory(node.NodeType, node, result.message, executionTime, result.success);

      return result;
    } catch (error) {
      this.executionState.failedNodes.add(node.NodeCode);
      this.addHistory(node.NodeType, node, error.message, new Date() - startTime, false);

      return {
        success: false,
        message: error.message,
      };
    }
  }

  /**
   * 执行开始节点
   * @param {Object} node - 开始节点
   * @returns {Object} 执行结果
   */
  async executeStartNode(node) {
    // 开始节点通常只需要记录启动信息
    this.executionState.variables.startTime = new Date().toISOString();
    this.executionState.variables.startUser = this.executionState.variables.currentUser || 'system';

    return {
      success: true,
      message: '流程启动成功',
    };
  }

  /**
   * 执行审批节点
   * @param {Object} node - 审批节点
   * @returns {Object} 执行结果
   */
  async executeApprovalNode(node) {
    // 根据 ApproverType 分发到不同的处理函数
    if (node.ApproverType === 'AUTO_APPROVE') {
      return this.executeAutoApproveNode(node);
    }
    if (node.ApproverType === 'AUTO_REJECT') {
      return this.executeAutoRejectNode(node);
    }

    // 默认处理标准审批（如 USERS）
    const approverConfig = JsonHelper.safeParse(node.ApproverConfig);

    if (!approverConfig.users || approverConfig.users.length === 0) {
      return {
        success: false,
        message: '审批节点未配置审批人',
      };
    }

    // 在实际应用中，这里会触发审批流程
    // 这里简化处理，直接通过审批
    const approvers = approverConfig.users.join(', ');

    // 模拟审批处理时间
    await this.sleep(100);

    this.executionState.variables.lastApprover = approvers[0];
    this.executionState.variables.approvalTime = new Date().toISOString();

    return {
      success: true,
      message: `审批通过，审批人: ${approvers}`,
    };
  }

  /**
   * 执行条件分支节点
   * @param {Object} node - 条件节点
   * @returns {Object} 执行结果
   */
  async executeConditionNode(node) {
    const conditionConfig = JsonHelper.safeParse(node.ConditionConfig);

    if (!conditionConfig.branches || conditionConfig.branches.length === 0) {
      return {
        success: false,
        message: '条件节点缺少有效分支配置',
      };
    }

    // 评估条件，选择匹配的分支
    const matchedBranch = await this.evaluateConditions(conditionConfig.branches);

    if (!matchedBranch) {
      return {
        success: false,
        message: '没有匹配的条件分支',
      };
    }

    // 执行匹配分支的流程
    if (matchedBranch.nodes && matchedBranch.nodes.length > 0) {
      const branchResult = await this.executeFlow(matchedBranch.nodes, matchedBranch.nodes[0]);

      if (!branchResult.success) {
        return {
          success: false,
          message: `条件分支执行失败: ${branchResult.message}`,
        };
      }
    }

    return {
      success: true,
      message: `条件分支执行完成，匹配分支: ${matchedBranch.name}`,
    };
  }

  /**
   * 执行抄送节点
   * @param {Object} node - 抄送节点
   * @returns {Object} 执行结果
   */
  async executeCCNode(node) {
    const approverConfig = JsonHelper.safeParse(node.ApproverConfig);

    if (!approverConfig.users || approverConfig.users.length === 0) {
      return {
        success: false,
        message: '抄送节点未配置抄送人',
      };
    }

    // 在实际应用中，这里会发送通知
    // 这里简化处理，记录抄送信息
    const ccUsers = approverConfig.users.join(', ');

    // 模拟发送通知时间
    await this.sleep(50);

    this.executionState.variables.ccUsers = approverConfig.users;
    this.executionState.variables.ccTime = new Date().toISOString();

    return {
      success: true,
      message: `抄送完成，抄送人: ${ccUsers}`,
    };
  }

  /**
   * 执行自动通过节点
   * @param {Object} node - 自动通过节点
   * @returns {Object} 执行结果
   */
  async executeAutoApproveNode(node) {
    this.executionState.variables.approveTime = new Date().toISOString();
    this.executionState.variables.approveReason = '系统自动通过';

    return {
      success: true,
      message: '自动通过处理完成',
    };
  }

  /**
   * 执行自动驳回节点
   * @param {Object} node - 自动驳回节点
   * @returns {Object} 执行结果
   */
  async executeAutoRejectNode(node) {
    this.executionState.variables.rejectTime = new Date().toISOString();
    this.executionState.variables.rejectReason = '系统自动驳回';

    return {
      success: true,
      message: '自动驳回处理完成',
    };
  }

  /**
   * 执行结束节点
   * @param {Object} node - 结束节点
   * @returns {Object} 执行结果
   */
  async executeEndNode(node) {
    this.executionState.variables.endTime = new Date().toISOString();
    this.executionState.variables.totalTime = new Date() - this.startTime;

    return {
      success: true,
      message: '流程结束',
    };
  }

  /**
   * 评估条件分支
   * @param {Array} branches - 分支列表
   * @returns {Object|null} 匹配的分支
   */
  async evaluateConditions(branches) {
    // 简化的条件评估逻辑
    // 在实际应用中，这里会根据流程变量评估复杂的条件表达式

    for (const branch of branches) {
      if (!branch.condition) {
        // 默认分支（无条件）
        return branch;
      }

      const { fieldName, operator, fieldValue } = branch.condition;

      if (!fieldName || !operator) {
        continue;
      }

      // 获取变量值
      const variableValue = this.executionState.variables[fieldName];

      // 评估条件
      if (this.evaluateCondition(variableValue, operator, fieldValue)) {
        return branch;
      }
    }

    // 如果没有匹配的条件，返回最后一个分支（默认分支）
    return branches[branches.length - 1];
  }

  /**
   * 评估单个条件
   * @param {*} actualValue - 实际值
   * @param {string} operator - 操作符
   * @param {*} expectedValue - 期望值
   * @returns {boolean} 条件是否满足
   */
  evaluateCondition(actualValue, operator, expectedValue) {
    switch (operator) {
      case 'eq':
        return actualValue == expectedValue;
      case 'ne':
        return actualValue != expectedValue;
      case 'gt':
        return Number(actualValue) > Number(expectedValue);
      case 'gte':
        return Number(actualValue) >= Number(expectedValue);
      case 'lt':
        return Number(actualValue) < Number(expectedValue);
      case 'lte':
        return Number(actualValue) <= Number(expectedValue);
      case 'contains':
        return String(actualValue).includes(String(expectedValue));
      case 'in':
        return Array.isArray(expectedValue) && expectedValue.includes(actualValue);
      default:
        return false;
    }
  }

  /**
   * 查找下一个节点
   * @param {Array} nodes - 节点列表
   * @param {number} currentIndex - 当前节点索引
   * @param {Object} currentNode - 当前节点
   * @returns {Object|null} 下一个节点
   */
  findNextNode(nodes, currentIndex, currentNode) {
    if (currentIndex >= nodes.length - 1) {
      return null;
    }

    // 返回下一个节点
    return nodes[currentIndex + 1];
  }

  /**
   * 添加执行历史记录
   * @param {string} nodeType - 节点类型
   * @param {Object} node - 节点对象
   * @param {string} message - 执行消息
   * @param {number} executionTime - 执行时间
   * @param {boolean} success - 是否成功
   */
  addHistory(nodeType, node, message, executionTime = 0, success = true) {
    this.executionState.history.push({
      timestamp: new Date().toISOString(),
      nodeType,
      nodeCode: node.NodeCode,
      nodeName: node.NodeName,
      message,
      executionTime,
      success,
    });
  }

  /**
   * 获取执行状态
   * @returns {Object} 当前执行状态
   */
  getExecutionState() {
    return {
      ...this.executionState,
      completedNodes: Array.from(this.executionState.completedNodes),
      failedNodes: Array.from(this.executionState.failedNodes),
    };
  }

  /**
   * 暂停工作流执行
   * @returns {Object} 暂停结果
   */
  pause() {
    if (this.executionState.status !== 'running') {
      return {
        success: false,
        message: '只能暂停正在运行的工作流',
      };
    }

    this.executionState.status = 'paused';
    return {
      success: true,
      message: '工作流已暂停',
    };
  }

  /**
   * 恢复工作流执行
   * @returns {Object} 恢复结果
   */
  resume() {
    if (this.executionState.status !== 'paused') {
      return {
        success: false,
        message: '只能恢复已暂停的工作流',
      };
    }

    this.executionState.status = 'running';
    return {
      success: true,
      message: '工作流已恢复',
    };
  }

  /**
   * 重置工作流状态
   */
  reset() {
    this.executionState = {
      currentNodeCode: null,
      completedNodes: new Set(),
      failedNodes: new Set(),
      variables: {},
      history: [],
      status: 'ready',
    };
    this.startTime = null;
    this.endTime = null;
  }

  /**
   * 异步延迟函数
   * @param {number} ms - 延迟毫秒数
   * @returns {Promise} Promise对象
   */
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

/**
 * 创建工作流引擎实例
 * @param {Object} workflow - 工作流对象
 * @returns {WorkflowEngine} 工作流引擎实例
 */
export function createWorkflowEngine(workflow) {
  return new WorkflowEngine(workflow);
}
