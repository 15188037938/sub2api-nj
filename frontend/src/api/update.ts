/**
 * Online Update API endpoints
 */

import { apiClient } from './client'

export interface UpdateStatus {
  current_commit: string
  latest_commit: string
  build_time: string
  is_up_to_date: boolean
  update_available: boolean
}

export interface UpdateApplyResult {
  message: string
  status: string
}

/**
 * Get current update status
 */
export async function getUpdateStatus(): Promise<UpdateStatus> {
  const { data } = await apiClient.get<UpdateStatus>('/admin/update/status')
  return data
}

/**
 * Apply available update
 */
export async function applyUpdate(): Promise<UpdateApplyResult> {
  const { data } = await apiClient.post<UpdateApplyResult>('/admin/update/apply')
  return data
}
