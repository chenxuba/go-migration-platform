import { useGet, usePost } from '~/utils/request'

export interface FaceCollectionStudentItem {
  id: string
  stuName: string
  avatarUrl?: string
  mobile?: string
  isCollect?: boolean
}

export interface FaceCollectionStudentPagedResult {
  list?: FaceCollectionStudentItem[]
  total?: number
  current?: number
  size?: number
}

export interface FaceCollectionProfile {
  studentId: string
  stuName?: string
  faceDescriptor?: number[]
  faceImage?: string
  updatedTime?: string
}

export function pageFaceCollectionStudentsApi(data: {
  pageRequestModel: {
    pageSize: number
    pageIndex: number
  }
  queryModel?: {
    searchKey?: string
  }
}) {
  return usePost<FaceCollectionStudentPagedResult>('/api/v1/face-collections/students/page', data)
}

export function getFaceCollectionProfileApi(params: { studentId: string | number }) {
  return useGet<FaceCollectionProfile>('/api/v1/face-collections/profile', params, {
    silentError: true,
  })
}

export function listFaceCollectionProfilesApi() {
  return useGet<FaceCollectionProfile[]>('/api/v1/face-collections/profiles')
}

export function saveFaceCollectionProfileApi(data: {
  studentId: string | number
  faceDescriptor: number[]
  faceImage: string
}) {
  return usePost<boolean>('/api/v1/face-collections/save', data)
}

export function deleteFaceCollectionProfileApi(data: { studentId: string | number }) {
  return usePost<boolean>('/api/v1/face-collections/delete', data)
}
