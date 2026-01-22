/**
 * 人员服务
 * 提供从主程序获取人员列表的功能
 * 主程序需要提供数据获取方法
 */

// 主程序数据提供者，需要由主程序设置
let dataProvider = null;

export const UserService = {
  /**
   * 设置主程序数据提供者
   * @param {Object} provider - 数据提供者对象
   * @param {Function} provider.getUserList - 获取人员列表的方法
   */
  setDataProvider(provider) {
    if (!provider || typeof provider.getUserList !== 'function') {
      throw new Error('数据提供者必须包含 getUserList 方法');
    }
    dataProvider = provider;
  },

  /**
   * 检查数据提供者是否已设置
   * @returns {boolean}
   */
  hasDataProvider() {
    return dataProvider !== null;
  },

  /**
   * 获取人员列表
   * @param {Object} options - 获取选项
   * @param {string} options.username - 根据用户名搜索（可选）
   * @param {number} options.limit - 返回记录数量限制，默认10条
   * @returns {Promise<Array>} 人员列表
   */
  async getUserList(options = {}) {
    try {
      if (!dataProvider) {
        throw new Error('数据提供者未设置，请先调用 setDataProvider 方法');
      }

      // 设置默认限制为10条记录
      const finalOptions = {
        limit: 10,
        ...options,
      };

      return await dataProvider.getUserList(finalOptions);
    } catch (error) {
      console.error('获取人员列表失败:', error);
      throw new Error('获取人员列表失败，请检查数据提供者配置');
    }
  },
};

/**
 * 用户选择器工具函数
 */
export const UserSelectorUtils = {
  /**
   * 格式化用户选择器选项
   * @param {Array} users - 用户列表
   * @returns {Array} 格式化后的选项
   */
  formatUserOptions(users) {
    return users.map(user => ({
      value: user.id,
      label: user.username,
      user: user,
    }));
  },

  /**
   * 获取用户选择器配置
   * @param {Object} options - 配置选项
   * @returns {Object} 选择器配置
   */
  getSelectorConfig(options = {}) {
    return {
      multiple: options.multiple !== false,
      placeholder: options.placeholder || '请选择人员',
      allowClear: options.allowClear !== false,
      showSearch: options.showSearch !== false,
      filterOption: options.filterOption || true,
      ...options,
    };
  },
};

// 默认导出
export default UserService;
