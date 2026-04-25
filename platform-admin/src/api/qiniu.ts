import request from '@/utils/request'

/**
 * 获取七牛云上传token
 */
export function getQiniuToken(params?: { tenantId?: string }) {
  return request({
    url: '/api/v1/qiniu/upload-token',
    method: 'get',
    params,
    silentError: true,
  } as any)
}

/**
 * 获取视频上传token
 */
export function getVideoUploadToken(params?: { tenantId?: string }) {
  return request({
    url: '/api/v1/qiniu/video-upload-token',
    method: 'get',
    params,
    silentError: true,
  } as any)
}
