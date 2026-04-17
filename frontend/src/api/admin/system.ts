/**
 * System API endpoints for admin operations
 */

import { apiClient } from '../client'

export interface ReleaseInfo {
  name: string
  body: string
  published_at: string
  html_url: string
}

export interface VersionInfo {
  current_version: string
  latest_version: string
  has_update: boolean
  release_info?: ReleaseInfo
  cached: boolean
  warning?: string
  build_type: string // "source" for manual builds, "release" for CI builds
}

/**
 * Get current version
 */
export async function getVersion(): Promise<{ version: string }> {
  const { data } = await apiClient.get<{ version: string }>('/admin/system/version')
  return data
}

/**
 * Check for updates
 * @param force - Force refresh from GitHub API
 */
export async function checkUpdates(force = false): Promise<VersionInfo> {
  const { data } = await apiClient.get<VersionInfo>('/admin/system/check-updates', {
    params: force ? { force: 'true' } : undefined
  })
  return data
}

export interface UpdateResult {
  message: string
  need_restart: boolean
}

export interface DeployConfig {
  enabled: boolean
  mode: string
  source_type?: string
  default_image: string
  allowed_image_prefix?: string
  archive_url?: string
  loaded_image?: string
  service_name: string
  compose_project_dir: string
  compose_file?: string
  docker_binary?: string
  compose_binary?: string
}

export interface DeployState {
  status: string
  requested_image?: string
  last_message?: string
  last_error?: string
  started_at?: number
  finished_at?: number
}

export interface DeployResult {
  message: string
  need_restart: boolean
  status: string
  image: string
  service_name: string
  compose_dir: string
  commands?: string[]
}

/**
 * Perform system update
 * Downloads and applies the latest version
 */
export async function performUpdate(): Promise<UpdateResult> {
  const { data } = await apiClient.post<UpdateResult>('/admin/system/update')
  return data
}

export async function getDeployConfig(): Promise<DeployConfig> {
  const { data } = await apiClient.get<DeployConfig>('/admin/system/deploy-config')
  return data
}

export async function updateDeployConfig(payload: DeployConfig): Promise<DeployConfig> {
  const { data } = await apiClient.put<DeployConfig>('/admin/system/deploy-config', payload)
  return data
}

export async function getDeployStatus(): Promise<DeployState> {
  const { data } = await apiClient.get<DeployState>('/admin/system/deploy-status')
  return data
}

export async function triggerDeploy(payload?: { image?: string; dry_run?: boolean }): Promise<DeployResult> {
  const { data } = await apiClient.post<DeployResult>('/admin/system/deploy', payload ?? {})
  return data
}

/**
 * Rollback to previous version
 */
export async function rollback(): Promise<UpdateResult> {
  const { data } = await apiClient.post<UpdateResult>('/admin/system/rollback')
  return data
}

/**
 * Restart the service
 */
export async function restartService(): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/admin/system/restart')
  return data
}

export const systemAPI = {
  getVersion,
  checkUpdates,
  performUpdate,
  getDeployConfig,
  updateDeployConfig,
  getDeployStatus,
  triggerDeploy,
  rollback,
  restartService
}

export default systemAPI
