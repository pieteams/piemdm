import { Configuration } from './generated';

/**
 * API 配置
 */
export const apiConfig = new Configuration({
  basePath: import.meta.env.VITE_API_BASE_URL ?? '',
  baseOptions: {
    headers: {
      'Content-Type': 'application/json',
    },
  },
});

/**
 * 更新 API 配置（用于动态设置 token）
 */
export function updateApiConfig(token?: string) {
  if (token) {
    apiConfig.baseOptions = {
      ...apiConfig.baseOptions,
      headers: {
        ...apiConfig.baseOptions?.headers,
        Authorization: `Bearer ${token}`,
      },
    };
  }
}
