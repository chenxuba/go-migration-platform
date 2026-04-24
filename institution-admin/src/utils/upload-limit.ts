export interface UploadLimitToken {
  maxSize?: number
  mimeTypes?: string
}

export function formatUploadSize(bytes?: number) {
  const value = Number(bytes || 0)
  if (value <= 0)
    return ''
  if (value >= 1024 * 1024)
    return `${Number((value / 1024 / 1024).toFixed(2))}MB`
  if (value >= 1024)
    return `${Number((value / 1024).toFixed(2))}KB`
  return `${value}B`
}

export function matchUploadMime(fileType: string, mimeTypes?: string) {
  const normalizedFileType = String(fileType || '').toLowerCase()
  const raw = String(mimeTypes || '').trim().toLowerCase()
  if (!raw)
    return true
  return raw.split(';').map(item => item.trim()).filter(Boolean).some((rule) => {
    if (rule.endsWith('/*'))
      return normalizedFileType.startsWith(rule.slice(0, -1))
    return normalizedFileType === rule
  })
}

export function validateUploadFileByToken(file: File, token: UploadLimitToken | undefined, label = '文件') {
  if (!token)
    return true
  if (token.mimeTypes && !matchUploadMime(file.type, token.mimeTypes))
    throw new Error(`${label}格式不符合当前租户云存储限制（${token.mimeTypes}）`)
  if (token.maxSize && file.size > token.maxSize)
    throw new Error(`${label}大小不能超过 ${formatUploadSize(token.maxSize)}`)
  return true
}

export function resolveUploadErrorMessage(error: any, fallback = '上传失败') {
  return error?.response?.data?.message || error?.message || fallback
}
