/**
 * API Endpoints Configuration
 *
 * API 端点配置
 */

export interface ApiEndpoints {
  baseURL: string;
  userLogin: string;
  userLogout: string;
  adminLogin: string;
  adminLogout: string;
  validateToken: string;
}

const env = (import.meta as any).env || {};
// Single source for the API origin. Empty means same origin, which is what the
// container image uses: nginx proxies /api/ to the backend.
const apiBase = String(env.VITE_BASE_API ?? '').replace(/\/+$/, '');

const apiConfig: Record<string, ApiEndpoints> = {
  development: {
    baseURL: `${apiBase}/api/v1`,
    userLogin: '/auth/login',
    userLogout: '/auth/logout',
    adminLogin: '/admin/auth/login',
    adminLogout: '/admin/auth/logout',
    validateToken: '/auth/validate',
  },

  production: {
    baseURL: `${apiBase}/api/v1`,
    userLogin: '/auth/login',
    userLogout: '/auth/logout',
    adminLogin: '/admin/auth/login',
    adminLogout: '/admin/auth/logout',
    validateToken: '/auth/validate',
  },
};

const currentEnv = (import.meta as any).env?.MODE || 'development';
export default apiConfig[currentEnv] as ApiEndpoints;
