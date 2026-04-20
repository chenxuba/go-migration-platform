import { useGet, usePost } from '~/utils/request'

export interface ClassroomItem {
  id: number
  uuid?: string
  version?: number
  instId?: number
  name: string
  address?: string
  enabled: boolean
  createTime?: string
  updateTime?: string
}

export interface ClassroomMutation {
  id?: number
  uuid?: string
  version?: number
  name: string
  address?: string
  enabled?: boolean
}

export function listClassroomsApi(params?: {
  enabledOnly?: boolean
  searchKey?: string
}) {
  return useGet<ClassroomItem[]>('/api/v1/classrooms', params)
}

export function createClassroomApi(data: ClassroomMutation) {
  return usePost<{ id: number }>('/api/v1/classrooms/create', data)
}

export function updateClassroomApi(data: ClassroomMutation) {
  return usePost<boolean>('/api/v1/classrooms/update', data)
}

export function updateClassroomStatusApi(data: {
  id: number
  enabled: boolean
}) {
  return usePost<boolean>('/api/v1/classrooms/status', data)
}

export function deleteClassroomApi(data: {
  id: number
}) {
  return usePost<boolean>('/api/v1/classrooms/delete', data)
}
