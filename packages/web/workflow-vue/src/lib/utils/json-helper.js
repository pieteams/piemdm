/**
 * JSON解析和序列化工具类
 * 提供安全的JSON操作方法，统一错误处理
 */
export const JsonHelper = {
  /**
   * 安全解析JSON字符串
   * @param {string} jsonString - 要解析的JSON字符串
   * @param {*} defaultValue - 解析失败时的默认值
   * @returns {*} 解析结果或默认值
   */
  safeParse(jsonString, defaultValue = {}) {
    if (!jsonString || jsonString === '{}') {
      return defaultValue;
    }

    try {
      return JSON.parse(jsonString);
    } catch (e) {
      console.warn('JSON解析失败:', e.message, '原始数据:', jsonString);
      return defaultValue;
    }
  },

  /**
   * 安全序列化对象为JSON字符串
   * @param {*} obj - 要序列化的对象
   * @param {string} defaultValue - 序列化失败时的默认值
   * @returns {string} JSON字符串或默认值
   */
  safeStringify(obj, defaultValue = '{}') {
    try {
      return JSON.stringify(obj);
    } catch (e) {
      console.warn('JSON序列化失败:', e.message, '原始对象:', obj);
      return defaultValue;
    }
  },

  /**
   * 检查字符串是否为有效的JSON
   * @param {string} jsonString - 要检查的字符串
   * @returns {boolean} 是否为有效JSON
   */
  isValid(jsonString) {
    try {
      JSON.parse(jsonString);
      return true;
    } catch (e) {
      return false;
    }
  },
};
