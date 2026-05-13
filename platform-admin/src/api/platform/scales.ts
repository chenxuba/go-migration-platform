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
  materialImages: string[]
  method: string
  describes?: string
  guidance: string
  guidanceVideo: string
  domainCode: string
  domainName: string
  standard: string
  scoreOptions: ScaleQuestionBankScoreOption[]
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

export interface PlatformPageRequestModel {
  pageIndex: number
  pageSize: number
}

export interface PlatformPageResult<T> {
  items: T[]
  total: number
  current: number
  size: number
}

export type PEP3IEPMaterialStatus = 'active' | 'inactive' | string

export interface PEP3IEPMaterialQuery {
  materialType?: 'long_term' | 'short_term' | string
  parentGoalMaterialId?: number
  goalMaterialId?: number
  domainCode?: string
  domain?: string
  courseForm?: string
  status?: PEP3IEPMaterialStatus
  keyword?: string
}

export interface PEP3IEPItemOptionRuleQuery {
  itemNo?: number
  scoreValue?: number
  domainCode?: string
  domain?: string
  status?: PEP3IEPMaterialStatus
  keyword?: string
}

export interface PEP3IEPItemOptionRulePageRequest {
  pageRequestModel: PlatformPageRequestModel
  queryModel: PEP3IEPItemOptionRuleQuery
}

export interface PEP3IEPMaterialPageRequest {
  pageRequestModel: PlatformPageRequestModel
  queryModel: PEP3IEPMaterialQuery
}

export interface PEP3IEPGoalMaterial {
  id?: number
  libraryScope: 'platform' | string
  instId?: number
  materialType?: 'long_term' | 'short_term' | string
  parentGoalMaterialId?: number
  domainCode?: string
  domain?: string
  longGoal: string
  shortGoal: string
  courseForm?: string
  ageMinMonths?: number
  ageMaxMonths?: number
  difficultyLevel?: number
  applicableScoreValues?: string
  priority?: number
  status?: PEP3IEPMaterialStatus
  createdTime?: string
  updatedTime?: string
}

export interface PEP3IEPItemOptionRule {
  id?: number
  libraryScope: 'platform' | string
  instId?: number
  itemNo: number
  itemTitle?: string
  domainCode?: string
  domain?: string
  scoreValue: number
  scoreLabel?: string
  scoreDescription?: string
  resultMeaning?: string
  generatePolicy?: string
  priority?: number
  aiInstruction?: string
  status?: PEP3IEPMaterialStatus
  goalMaterialIds?: number[]
  goalMaterials?: PEP3IEPGoalMaterial[]
  createdTime?: string
  updatedTime?: string
}

export interface PEP3IEPTrainingMaterial {
  id?: number
  libraryScope: 'platform' | string
  instId?: number
  goalMaterialId?: number
  trainingProject: string
  trainingContent: string
  priority?: number
  status?: PEP3IEPMaterialStatus
  createdTime?: string
  updatedTime?: string
}

export function pagePlatformPEP3IEPMaterialRulesApi(data: PEP3IEPItemOptionRulePageRequest) {
  return usePost<PlatformPageResult<PEP3IEPItemOptionRule>>('/api/v1/platform/scales/pep3-iep-material/rules/page', data, { silentError: true })
}

export function savePlatformPEP3IEPMaterialRuleApi(data: PEP3IEPItemOptionRule) {
  return usePost<PEP3IEPItemOptionRule>('/api/v1/platform/scales/pep3-iep-material/rules/save', data)
}

export function deletePlatformPEP3IEPMaterialRuleApi(id: number) {
  return usePost<{ deleted: boolean }>('/api/v1/platform/scales/pep3-iep-material/rules/delete', { id })
}

export function pagePlatformPEP3IEPMaterialGoalsApi(data: PEP3IEPMaterialPageRequest) {
  return usePost<PlatformPageResult<PEP3IEPGoalMaterial>>('/api/v1/platform/scales/pep3-iep-material/goals/page', data, { silentError: true })
}

export function savePlatformPEP3IEPMaterialGoalApi(data: PEP3IEPGoalMaterial) {
  return usePost<PEP3IEPGoalMaterial>('/api/v1/platform/scales/pep3-iep-material/goals/save', data)
}

export function deletePlatformPEP3IEPMaterialGoalApi(id: number) {
  return usePost<{ deleted: boolean }>('/api/v1/platform/scales/pep3-iep-material/goals/delete', { id })
}

export function pagePlatformPEP3IEPMaterialTrainingApi(data: PEP3IEPMaterialPageRequest) {
  return usePost<PlatformPageResult<PEP3IEPTrainingMaterial>>('/api/v1/platform/scales/pep3-iep-material/training/page', data, { silentError: true })
}

export function savePlatformPEP3IEPMaterialTrainingApi(data: PEP3IEPTrainingMaterial) {
  return usePost<PEP3IEPTrainingMaterial>('/api/v1/platform/scales/pep3-iep-material/training/save', data)
}

export function deletePlatformPEP3IEPMaterialTrainingApi(id: number) {
  return usePost<{ deleted: boolean }>('/api/v1/platform/scales/pep3-iep-material/training/delete', { id })
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
