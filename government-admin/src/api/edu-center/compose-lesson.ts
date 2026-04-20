import { usePost } from '~/utils/request'

/** 对标 GetPageComposeLessonListForPc 列表项 */
export interface ComposeLessonListItem {
  id: string
  name: string
  createTime: string
  productCount: number
  classCount: number
}

/** 对标 ToB/PC/Lesson/Create */
export function createComposeLessonApi(data: {
  lessonName: string
  productIds: string[]
}) {
  return usePost<{ id: string; name: string }>('/api/v1/compose-lessons/create', data)
}

/** 对标 GetPageComposeLessonListForPc */
export function pageComposeLessonsForPcApi(data: {
  queryModel: { searchKey: string }
  pageRequestModel: {
    needTotal?: boolean
    skipCount?: number
    pageSize?: number
    pageIndex?: number
  }
}) {
  return usePost<{ list: ComposeLessonListItem[]; total: number }>(
    '/api/v1/compose-lessons/page',
    data,
  )
}
