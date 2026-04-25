import type { ResponseBody } from '@/utils/request'

export type LoginTemplateEntryType = 'platform-admin' | 'institution-admin' | 'all'
export type LoginTemplateLayoutType = 'split' | 'card' | 'portal'

export interface LoginTemplateItem {
  id: number
  templateKey: string
  templateName: string
  entryType: LoginTemplateEntryType | string
  layoutType?: LoginTemplateLayoutType | string
  description?: string
  previewImage?: string
  enabled: boolean
  sort: number
  tenantIds?: string[]
  institutionIds?: number[]
  referenceCount: number
  createTime?: string
  updateTime?: string
}

export interface LoginTemplateMutationPayload {
  id?: number
  templateKey: string
  templateName: string
  entryType: LoginTemplateEntryType | string
  layoutType?: LoginTemplateLayoutType | string
  description?: string
  previewImage?: string
  enabled?: boolean
  sort?: number
  tenantIds?: string[]
  institutionIds?: number[]
}

export function listLoginTemplatesApi(params: { entryType?: string, tenantId?: string, institutionId?: number, enabledOnly?: boolean } = {}) {
  return useGet<LoginTemplateItem[], typeof params>('/api/v1/platform/login-templates', params) as Promise<ResponseBody<LoginTemplateItem[]>>
}

export function saveLoginTemplateApi(data: LoginTemplateMutationPayload) {
  return usePost<{ id: number }, LoginTemplateMutationPayload>('/api/v1/platform/login-templates/save', data)
}

export function deleteLoginTemplateApi(data: { id: number }) {
  return usePost<boolean, { id: number }>('/api/v1/platform/login-templates/delete', data)
}
