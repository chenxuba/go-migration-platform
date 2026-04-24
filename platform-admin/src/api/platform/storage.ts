export interface TenantStorageConfig {
  tenantId: string
  provider: 'qiniu'
  accessKey: string
  secretKey?: string
  bucket: string
  bucketHost: string
  uploadPrefix?: string
  expiresSeconds?: number
  imageMaxSize?: number
  imageMimeTypes?: string
  videoMaxSize?: number
  videoMimeTypes?: string
  enabled: boolean
  remark?: string
  updateTime?: string
}

export function getTenantStorageConfigApi(params: { tenantId?: string } = {}) {
  return useGet<TenantStorageConfig, { tenantId?: string }>('/api/v1/platform/tenant-storage', params, { silentError: true })
}

export function saveTenantStorageConfigApi(data: TenantStorageConfig) {
  return usePost<boolean, TenantStorageConfig>('/api/v1/platform/tenant-storage', data, { silentError: true })
}
