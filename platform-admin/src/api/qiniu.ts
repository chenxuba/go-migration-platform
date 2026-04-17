import request from '@/utils/request'

/**
 * 获取七牛云上传token
 */
export function getQiniuToken() {
  return request({
    url: '/api/v1/qiniu/upload-token',
    method: 'get'
  })
}

/**
 * 获取视频上传token
 */
export function getVideoUploadToken() {
  return request({
    url: '/api/v1/qiniu/video-upload-token',
    method: 'get'
  })
}
