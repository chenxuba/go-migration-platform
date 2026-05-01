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
  ageMinMonths: number
  ageMaxMonths: number
  duration: string
  durationMinMinutes: number
  durationMaxMinutes: number
  currentVersion: string
  itemCount: number
  domainCount: number
  institutionCount: number
  monthUsage: number
  dataStatus: string
  updatedAt: string
  summary: string
  posterUrl: string
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

export interface ScaleQuestionBankDomain {
  scaleCode: string
  scaleName: string
  category?: string
  itemCount?: number
  maxRawScore?: number
  itemNumbers?: number[]
  isDevelopmentSubtest?: boolean
  isBehaviorSubtest?: boolean
  isCaregiverReport?: boolean
  compositeCode?: string
}

export interface ScaleQuestionBankScoreOption {
  value: number
  label: string
  description?: string
}

export interface ScaleQuestionBankRecordFieldOption {
  value: string
  label: string
}

export interface ScaleQuestionBankRecordField {
  key: string
  label: string
  fieldType: string
  displayType?: string
  required?: boolean
  placeholder?: string
  options?: ScaleQuestionBankRecordFieldOption[]
}

export interface ScaleQuestionBankItem {
  itemNo: number
  itemTitle: string
  testItem: string
  materials: string
  method: string
  describes?: string
  guidance: string
  domainCode: string
  domainName: string
  standard: string
  scoreOptions: ScaleQuestionBankScoreOption[]
  scoreOptionText: string
  recordFields: ScaleQuestionBankRecordField[]
  sourcePdf?: string
  sourcePages?: number[]
  ocrStatus?: string
  updatedAt?: string
}

export interface ScaleQuestionBank {
  scaleCode: string
  scaleVersion: string
  dataStatus: string
  itemCount: number
  domainCount: number
  domains: ScaleQuestionBankDomain[]
  items: ScaleQuestionBankItem[]
  sourceTables: string[]
}

export interface ScaleQuestionBankParams {
  scaleCode: string
  scaleVersion?: string
}

export function getScaleQuestionBankApi(params: ScaleQuestionBankParams) {
  return useGet<ScaleQuestionBank, ScaleQuestionBankParams>('/api/v1/platform/scales/question-bank', params, { silentError: true })
}

export function updateScaleQuestionBankItemApi(data: ScaleQuestionBankItem & { scaleCode: string, scaleVersion: string }) {
  return usePost<boolean, ScaleQuestionBankItem & { scaleCode: string, scaleVersion: string }>('/api/v1/platform/scales/question-bank/items/update', data, { silentError: true })
}

export interface ScaleMutationPayload {
  id?: number
  name: string
  code?: string
  category: string
  scenario: string
  ageRange: string
  ageMinMonths: number
  ageMaxMonths: number
  currentVersion: string
  itemCount: number
  domainCount: number
  summary?: string
  posterUrl?: string
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
