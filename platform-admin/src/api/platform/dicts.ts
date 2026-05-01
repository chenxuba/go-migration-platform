import type { ResponseBody } from '@/utils/request'

export interface DictItem {
  id: number
  dictName: string
  dictCode: string
  isEnable: boolean
  remark?: string
}

export interface DictValueItem {
  id: number
  dictId: number
  dictLabel: string
  dictValue: string
  sort: number
  isEnable: boolean
}

export interface DictPageParams {
  current?: number
  size?: number
  keyword?: string
  scope?: string
}

export interface DictMutationPayload {
  id?: number
  dictName: string
  dictCode: string
  isEnable: boolean
  remark?: string
}

export interface DictValueMutationPayload {
  id?: number
  dictId?: number
  dictLabel: string
  dictValue: string
  sort: number
  isEnable: boolean
  remark?: string
}

export interface DictPagePayload {
  items: DictItem[]
  total: number
  current: number
  size: number
}

export function pageDictsApi(params: DictPageParams) {
  return useGet<DictItem[], DictPageParams>('/api/v1/platform/dicts', params) as Promise<
    ResponseBody<DictItem[]> & { data?: DictPagePayload }
  >
}

export function createDictApi(data: DictMutationPayload) {
  return usePost<{ id: number }, DictMutationPayload>('/api/v1/platform/dicts/create', data, { silentError: true })
}

export function updateDictApi(data: DictMutationPayload & { id: number }) {
  return usePost<boolean, DictMutationPayload & { id: number }>('/api/v1/platform/dicts/update', data, { silentError: true })
}

export function deleteDictApi(data: { id: number }) {
  return usePost<boolean, { id: number }>('/api/v1/platform/dicts/delete', data, { silentError: true })
}

export function listDictValuesApi(params: { code: string }) {
  return useGet<DictValueItem[], { code: string }>('/api/v1/platform/dict-values', params)
}

export function createDictValueApi(data: DictValueMutationPayload & { dictId: number }) {
  return usePost<{ id: number }, DictValueMutationPayload & { dictId: number }>(
    '/api/v1/platform/dict-values/create',
    data,
    { silentError: true },
  )
}

export function updateDictValueApi(data: DictValueMutationPayload & { id: number }) {
  return usePost<boolean, DictValueMutationPayload & { id: number }>(
    '/api/v1/platform/dict-values/update',
    data,
    { silentError: true },
  )
}

export function deleteDictValueApi(data: { id: number }) {
  return usePost<boolean, { id: number }>('/api/v1/platform/dict-values/delete', data, { silentError: true })
}
