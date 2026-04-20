import { useGet } from '~/utils/request'

export interface NoticeTemplateItem {
  id: string
  title: string
  coverUrl: string
  tag: string
  weight: number
  content: string
  summary: string
  orgId: string
  schoolId: string
}

export function listNoticeTemplatesApi() {
  return useGet<NoticeTemplateItem[]>('/api/v1/notices/templates')
}
