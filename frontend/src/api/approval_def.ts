/**
 * Approval Definition API
 *
 * 审批定义管理相关 API 封装
 *
 * 注意: 由于后端未添加 Swagger 注释,这里使用手动封装的 TypeScript 版本
 * TODO: 后端添加 Swagger 注释后,可以使用自动生成的 API 客户端
 */

import requestService from '@/utils/request';
import type { AxiosInstance, AxiosResponse } from 'axios';
import type { ApiResponse } from '@/api/types';

// 类型断言: request.js 导出的 service 是一个 Axios 实例
const service = requestService as AxiosInstance;

/**
 * ApprovalDefinition 数据模型
 */
export interface ApprovalDefinition {
  ID?: number;
  CreatedAt?: string;
  UpdatedAt?: string;
  DeletedAt?: string | null;
  Code?: string;
  Name: string;
  FormData?: string;
  NodeList?: string;
  Description?: string;
  Platform?: string;
  Status?: 'Normal' | 'Frozen' | 'Deleted';
  CreatedBy?: string;
  UpdatedBy?: string;
}

/**
 * 创建 ApprovalDefinition 请求参数
 */
export interface CreateApprovalDefRequest {
  Code?: string;
  Name: string;
  FormData?: string;
  NodeList?: string;
  Description?: string;
  Platform?: string;
  Status?: 'Normal' | 'Frozen' | 'Deleted';
}

/**
 * 更新 ApprovalDefinition 请求参数
 */
export interface UpdateApprovalDefRequest {
  id: number;
  Code?: string;
  Name?: string;
  FormData?: string;
  NodeList?: string;
  Description?: string;
  Platform?: string;
  Status?: 'Normal' | 'Frozen' | 'Deleted';
}

/**
 * 查询参数
 */
export interface GetApprovalDefListParams {
  page?: number;
  pageSize?: number;
  showType?: string;
  startDate?: string;
  endDate?: string;
  Code?: string;
  Name?: string;
  Status?: string;
}

/**
 * 查询单个 ApprovalDefinition 参数
 */
export interface GetApprovalDefParams {
  id: number;
}

/**
 * 创建 ApprovalDefinition
 *
 * @param data - 创建参数
 * @returns Promise<AxiosResponse<ApiResponse<ApprovalDefinition>>>
 *
 * @example
 * ```typescript
 * const response = await createApprovalDef({
 *   Name: 'My Approval',
 *   Description: 'Test approval definition',
 *   Status: 'Normal'
 * });
 * ```
 */
export const createApprovalDef = (
  data: CreateApprovalDefRequest
): Promise<AxiosResponse<ApiResponse<ApprovalDefinition>>> => {
  return service.post('/admin/approval_defs', data);
};

/**
 * 删除 ApprovalDefinition
 *
 * @param data - 包含 id 的对象
 * @returns Promise<AxiosResponse<ApiResponse>>
 */
export const deleteApprovalDef = (
  data: { id: number }
): Promise<AxiosResponse<ApiResponse>> => {
  return service.delete(`/admin/approval_defs/${data.id}`);
};

/**
 * 批量删除 ApprovalDefinition
 *
 * @param data - 批量删除参数
 * @returns Promise<AxiosResponse<ApiResponse>>
 */
export const batchDeleteApprovalDef = (
  data: { ids: number[]; action?: string }
): Promise<AxiosResponse<ApiResponse>> => {
  return service.delete('/admin/approval_defs/batch', { data });
};

/**
 * 更新 ApprovalDefinition
 *
 * @param data - 更新参数
 * @returns Promise<AxiosResponse<ApiResponse>>
 */
export const updateApprovalDef = (
  data: UpdateApprovalDefRequest
): Promise<AxiosResponse<ApiResponse>> => {
  return service.put(`/admin/approval_defs/${data.id}`, data);
};

/**
 * 批量更新 ApprovalDefinition 状态
 *
 * @param data - 批量更新参数
 * @returns Promise<AxiosResponse<ApiResponse>>
 */
export const updateApprovalDefStatus = (
  data: { ids: number[]; status: 'Normal' | 'Frozen' | 'Deleted' }
): Promise<AxiosResponse<ApiResponse>> => {
  return service.put('/admin/approval_defs/batch', data);
};

/**
 * 根据 ID 查询 ApprovalDefinition
 *
 * @param params - 查询参数
 * @returns Promise<AxiosResponse<ApiResponse<ApprovalDefinition>>>
 */
export const findApprovalDef = (
  params: GetApprovalDefParams
): Promise<AxiosResponse<ApiResponse<ApprovalDefinition>>> => {
  return service.get(`/admin/approval_defs/${params.id}`);
};

/**
 * 分页获取 ApprovalDefinition 列表
 *
 * @param params - 查询参数
 * @returns Promise<AxiosResponse<ApiResponse<ApprovalDefinition[]>>>
 */
export const getApprovalDefList = (
  params?: GetApprovalDefListParams
): Promise<AxiosResponse<ApiResponse<ApprovalDefinition[]>>> => {
  return service.get('/admin/approval_defs', { params });
};
