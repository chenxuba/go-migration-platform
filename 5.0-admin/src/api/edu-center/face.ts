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
  faceImage?: string
  updatedTime?: string
}

export interface FaceAttendanceRecord {
  id: string
  studentId: string
  studentName?: string
  faceImage?: string
  recordTime?: string
}

export interface FaceCompareResult {
  matched: boolean
  studentId?: string
  studentName?: string
  distance?: number
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

export function compareFaceCollectionApi(data: {
  faceDescriptor: number[]
}) {
  return usePost<FaceCompareResult>('/api/v1/face-collections/compare', data)
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

export function listFaceAttendanceRecordsApi(params?: { limit?: number }) {
  return useGet<FaceAttendanceRecord[]>('/api/v1/face-collections/attendance-records', params)
}

export function saveFaceAttendanceRecordApi(data: {
  studentId: string | number
  faceImage: string
}) {
  return usePost<FaceAttendanceRecord>('/api/v1/face-collections/attendance-records/save', data)
}
