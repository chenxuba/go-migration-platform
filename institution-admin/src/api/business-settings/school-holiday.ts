import { useGet, usePost } from '~/utils/request'

export interface SchoolHolidayItem {
  id: number
  instId?: number
  name: string
  startDate: string
  endDate: string
  source: 'statutory' | 'custom'
  sort?: number
}

export interface SchoolHolidayMutation {
  id?: number
  name: string
  startDate: string
  endDate: string
  source?: 'statutory' | 'custom'
  sort?: number
}

export function listSchoolHolidaysApi() {
  return useGet<SchoolHolidayItem[]>('/api/v1/school-holidays')
}

export function saveSchoolHolidayApi(data: SchoolHolidayMutation) {
  return usePost<{ id: number }>('/api/v1/school-holidays/save', data)
}

export function deleteSchoolHolidayApi(data: { id: number }) {
  return usePost<boolean>('/api/v1/school-holidays/delete', data)
}

export function resetSchoolHolidaysApi() {
  return usePost<boolean>('/api/v1/school-holidays/reset', {})
}
