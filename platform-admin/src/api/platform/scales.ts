export interface ScaleInstitutionRow {
  name: string
  contact: string
  authState: string
  expireAt: string
}

export interface ScaleTextResourceItem {
  id: number
  scaleId: number
  content: string
  sort: number
}

export interface ScaleRecord {
  id: number
  name: string
  code: string
  category: string
  scenario: string
  ageRange: string
  currentVersion: string
  itemCount: number
  domainCount: number
  institutionCount: number
  monthUsage: number
  dataStatus: string
  updatedAt: string
  summary: string
  executionEntry: string
  apiPackage: string
  references: ScaleTextResourceItem[]
  acknowledgements: ScaleTextResourceItem[]
  authInstitutions: ScaleInstitutionRow[]
}

export interface ScaleListParams {
  keyword?: string
  category?: string
  scenario?: string
}

export function listScalesApi(params?: ScaleListParams) {
  return useGet<ScaleRecord[], ScaleListParams>('/api/v1/platform/scales', params)
}

export interface ScaleMutationPayload {
  id?: number
  name: string
  code?: string
  category: string
  scenario: string
  ageRange: string
  currentVersion: string
  itemCount: number
  domainCount: number
  summary?: string
  executionEntry?: string
  apiPackage?: string
}

export function createScaleApi(data: ScaleMutationPayload) {
  return usePost<{ id: number }, ScaleMutationPayload>('/api/v1/platform/scales/create', data, { silentError: true })
}

export function updateScaleApi(data: ScaleMutationPayload & { id: number }) {
  return usePost<boolean, ScaleMutationPayload & { id: number }>('/api/v1/platform/scales/update', data, { silentError: true })
}

export interface ScaleTextResourceMutationPayload {
  id?: number
  scaleId?: number
  content: string
  sort?: number
}

export function createScaleReferenceApi(data: ScaleTextResourceMutationPayload) {
  return usePost<{ id: number }, ScaleTextResourceMutationPayload>('/api/v1/platform/scales/references/create', data, { silentError: true })
}

export function updateScaleReferenceApi(data: ScaleTextResourceMutationPayload & { id: number }) {
  return usePost<boolean, ScaleTextResourceMutationPayload & { id: number }>('/api/v1/platform/scales/references/update', data, { silentError: true })
}

export function deleteScaleReferenceApi(data: { id: number }) {
  return usePost<boolean, { id: number }>('/api/v1/platform/scales/references/delete', data, { silentError: true })
}

export function createScaleAcknowledgementApi(data: ScaleTextResourceMutationPayload) {
  return usePost<{ id: number }, ScaleTextResourceMutationPayload>('/api/v1/platform/scales/acknowledgements/create', data, { silentError: true })
}

export function updateScaleAcknowledgementApi(data: ScaleTextResourceMutationPayload & { id: number }) {
  return usePost<boolean, ScaleTextResourceMutationPayload & { id: number }>('/api/v1/platform/scales/acknowledgements/update', data, { silentError: true })
}

export function deleteScaleAcknowledgementApi(data: { id: number }) {
  return usePost<boolean, { id: number }>('/api/v1/platform/scales/acknowledgements/delete', data, { silentError: true })
}
