import type { LoginTemplateScope } from './login-template-registry'

interface RealLoginTemplatePreviewOptions {
  scope: LoginTemplateScope
  template: string
  name?: string
  desc?: string
  layout?: string
}

function resolveInstitutionAdminPreviewBase() {
  const { protocol, hostname, port } = window.location
  if (hostname === 'localhost' || hostname.endsWith('.localhost')) {
    const previewPort = port === '6688' ? '6678' : port
    return `${protocol}//${hostname}${previewPort ? `:${previewPort}` : ''}/`
  }
  return `${window.location.origin}${window.location.pathname}`
}

function resolvePlatformAdminPreviewBase() {
  return `${window.location.origin}${window.location.pathname}`
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
  const basePath = options.scope === 'institution' ? resolveInstitutionAdminPreviewBase() : resolvePlatformAdminPreviewBase()
  return `${basePath}#/login-template-preview?${params.toString()}`
}

export function openRealLoginTemplatePreview(options: RealLoginTemplatePreviewOptions) {
  window.open(buildRealLoginTemplatePreviewUrl(options), '_blank', 'noopener,noreferrer')
}
