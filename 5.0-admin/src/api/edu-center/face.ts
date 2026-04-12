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

export interface FaceAttendanceRelatedScheduleItem {
  scheduleId?: string
  classTime?: string
  scheduleName?: string
  rollCallStatus?: string
}

export interface FaceAttendanceRecordItem {
  id: string
  sessionId: string
  studentId: string
  studentName?: string
  studentMobile?: string
  avatarUrl?: string
  studentSex?: number
  isCollect?: boolean
  attendanceDate?: string
  sessionStatus?: number
  attendanceType?: string
  action?: string
  actionLabel?: string
  signInImage?: string
  signOutImage?: string
  attendanceTime?: string
  actionTime?: string
  signOutTime?: string
  hasSchedule?: boolean
  classTimes?: string[]
  relatedSchedules?: string[]
  relatedScheduleItems?: FaceAttendanceRelatedScheduleItem[]
  prompt?: string
}

export interface FaceCompareResult {
  matched: boolean
  studentId?: string
  studentName?: string
  distance?: number
}

export interface FaceAttendanceSession {
  id: string
  studentId: string
  studentName?: string
  avatarUrl?: string
  attendanceDate?: string
  status?: number
  signInTime?: string
  signInImage?: string
  signOutTime?: string
  signOutImage?: string
  latestAction?: string
  latestActionLabel?: string
  latestTime?: string
  latestImage?: string
  hasSchedule?: boolean
  lastLessonEndTime?: string
}

export interface FaceAttendanceRecognizeResult {
  matched: boolean
  studentId?: string
  studentName?: string
  avatarUrl?: string
  distance?: number
  action?: string
  actionLabel?: string
  sessionId?: string
  needUpload?: boolean
  message?: string
  hasSchedule?: boolean
  lastActionTime?: string
  lastLessonEndTime?: string
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

export function listFaceAttendanceSessionsApi(params?: { limit?: number }) {
  return useGet<FaceAttendanceSession[]>('/api/v1/face-collections/attendance-sessions', params)
}

export function recognizeFaceAttendanceSessionApi(data: {
  faceDescriptor: number[]
}) {
  return usePost<FaceAttendanceRecognizeResult>('/api/v1/face-collections/attendance-sessions/recognize', data)
}

export function commitFaceAttendanceSessionApi(data: {
  studentId: string | number
  sessionId?: string | number
  action: string
  faceImage: string
}) {
  return usePost<FaceAttendanceSession>('/api/v1/face-collections/attendance-sessions/commit', data)
}

export function listFaceAttendanceRecordsApi(params?: { limit?: number }) {
  return useGet<FaceAttendanceRecord[]>('/api/v1/face-collections/attendance-records', params)
}

export function pageFaceAttendanceRecordsApi(data: {
  pageRequestModel: {
    pageSize: number
    pageIndex: number
  }
  queryModel?: {
    studentId?: string | number
    actionTypes?: string[]
    beginSignInTime?: string
    endSignInTime?: string
    beginSignOutTime?: string
    endSignOutTime?: string
    pendingSignOut?: boolean
  }
}) {
  return usePost<FaceAttendanceRecordItem[]>('/api/v1/face-collections/attendance-records/page', data)
}

export function saveFaceAttendanceRecordApi(data: {
  studentId: string | number
  faceImage: string
}) {
  return usePost<FaceAttendanceRecord>('/api/v1/face-collections/attendance-records/save', data)
}
