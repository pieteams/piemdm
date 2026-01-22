/**
 * 用户服务测试文件
 * 测试用户服务的集成和使用
 */

import { describe, it, expect, vi } from 'vitest';
import { UserService, UserSelectorUtils } from '../src/lib/services/user-service.js';

describe('UserService Integration Tests', () => {
  // 设置默认的mock数据提供者用于测试
  const mockDataProvider = {
    async getUserList(options = {}) {
      const mockUsers = [
        { id: 6, username: 'jasen8888' },
        { id: 7, username: 'jasen' },
        { id: 8, username: 'testuser1' },
        { id: 9, username: 'testuser2' },
        { id: 10, username: 'admin' },
      ];

      let filteredUsers = [...mockUsers];

      // 根据用户名搜索
      if (options.username) {
        const searchTerm = options.username.toLowerCase();
        filteredUsers = filteredUsers.filter(user =>
          user.username.toLowerCase().includes(searchTerm),
        );
      }

      // 根据限制数量返回
      const limit = options.limit || 10;
      return filteredUsers.slice(0, limit);
    },
  };

  it('should set data provider successfully', () => {
    UserService.setDataProvider(mockDataProvider);
    expect(UserService.hasDataProvider()).toBe(true);
  });

  it('should get user list from data provider', async () => {
    UserService.setDataProvider(mockDataProvider);
    const users = await UserService.getUserList();
    expect(users).toHaveLength(5);
    expect(users[0].username).toBe('jasen8888');
  });

  it('should filter users by search keyword', async () => {
    UserService.setDataProvider(mockDataProvider);
    const searchResults = await UserService.getUserList({ username: 'jasen' });
    expect(searchResults).toHaveLength(2);
    expect(searchResults[0].username).toBe('jasen8888');
  });

  it('should search users by username using getUserList', async () => {
    UserService.setDataProvider(mockDataProvider);
    const searchResults = await UserService.getUserList({ username: 'test' });
    expect(searchResults).toHaveLength(2);
    expect(searchResults[0].username).toBe('testuser1');
  });

  it('should handle errors gracefully', async () => {
    const errorDataProvider = {
      async getUserList() {
        throw new Error('Network error');
      },
    };

    UserService.setDataProvider(errorDataProvider);

    try {
      await UserService.getUserList();
      // 如果代码执行到这里，说明没有抛出错误，测试应该失败
      expect(true).toBe(false);
    } catch (error) {
      expect(error.message).toBe('获取人员列表失败，请检查数据提供者配置');
    }
  });

  it('should use UserSelectorUtils correctly', () => {
    const users = [
      { id: 6, username: 'jasen8888' },
      { id: 7, username: 'jasen' },
    ];

    const options = UserSelectorUtils.formatUserOptions(users);
    expect(options).toHaveLength(2);
    expect(options[0].value).toBe(6);
    expect(options[0].label).toBe('jasen8888');

    const config = UserSelectorUtils.getSelectorConfig({ multiple: true });
    expect(config.multiple).toBe(true);
    expect(config.placeholder).toBe('请选择人员');
  });

  it('should work with limit option', async () => {
    UserService.setDataProvider(mockDataProvider);
    const limitedUsers = await UserService.getUserList({ limit: 2 });
    expect(limitedUsers).toHaveLength(2);
  });
});
