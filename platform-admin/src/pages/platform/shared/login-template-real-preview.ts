import type { LoginTemplateScope } from './login-template-registry'

interface RealLoginTemplatePreviewOptions {
  scope: LoginTemplateScope
  template: string
  name?: string
  desc?: string
  layout?: string
}

function resolveInstitutionAdminPreviewUrl() {
  const { protocol, hostname, port } = window.location
  if (hostname === 'localhost' || hostname.endsWith('.localhost')) {
    const previewPort = port === '6688' ? '6678' : port
    return `${protocol}//${hostname}${previewPort ? `:${previewPort}` : ''}/login-template-preview`
  }
  return `${window.location.origin}/institution/login-template-preview`
}

function resolvePlatformAdminPreviewUrl() {
  return `${window.location.origin}/platform/login-template-preview`
}

export function buildRealLoginTemplatePreviewUrl(options: RealLoginTemplatePreviewOptions) {
  const entryType = options.scope === 'institution' ? 'institution-admin' : 'platform-admin'
  const params = new URLSearchParams({
    template: options.template,
    entryType,
    name: options.name || (options.scope === 'institution' ? '机构端登录' : '客户管理后台'),
    layout: options.layout || 'split',
    desc: options.desc || '',
    templatePreview: '1',
  })
  const previewUrl = options.scope === 'institution' ? resolveInstitutionAdminPreviewUrl() : resolvePlatformAdminPreviewUrl()
  return `${previewUrl}?${params.toString()}`
}

export function openRealLoginTemplatePreview(options: RealLoginTemplatePreviewOptions) {
  window.open(buildRealLoginTemplatePreviewUrl(options), '_blank', 'noopener,noreferrer')
}
