import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { useGet, usePost } from '~/utils/request'

export type PEP3DateString = string
export type PEP3ScaleCode =
  | 'CVP'
  | 'EL'
  | 'RL'
  | 'FM'
  | 'GM'
  | 'VMI'
  | 'AE'
  | 'SR'
  | 'CMB'
  | 'CVB'
  | 'PB'
  | 'PSC'
  | 'AB'
  | string

export interface PageRequestModel {
  pageIndex: number
  pageSize: number
  skipCount?: number
}

export interface PageResult<T> {
  items: T[]
  total: number
  current: number
  size: number
}

export interface PEP3ItemScoreInput {
  itemNo: number
  score: 0 | 1 | 2 | number
}

export interface PEP3ItemRecordValueInput {
  itemNo: number
  fieldKey: string
  value: unknown
}

export interface PEP3RawScoreInput {
  scaleCode: PEP3ScaleCode
  rawScore: number
}

export interface PEP3ScoreRequest {
  birthDate: PEP3DateString
  assessmentDate: PEP3DateString
  itemScores?: Record<number, number>
  itemScoreList?: PEP3ItemScoreInput[]
  rawScores?: Partial<Record<PEP3ScaleCode, number>>
  rawScoreList?: PEP3RawScoreInput[]
  allowMissingItems?: boolean
}

export interface PEP3DraftSaveRequest {
  id?: number
  studentId?: number
  studentName?: string
  examinerName?: string
  remark?: string
  birthDate?: PEP3DateString
  assessmentDate?: PEP3DateString
  itemScores?: Record<number, number>
  itemScoreList?: PEP3ItemScoreInput[]
  rawScores?: Partial<Record<PEP3ScaleCode, number>>
  rawScoreList?: PEP3RawScoreInput[]
  itemRecordValues?: Record<number, Record<string, unknown>>
  itemRecordValueList?: PEP3ItemRecordValueInput[]
  allowMissingItems?: boolean
}

export interface PEP3RecordCreateRequest extends PEP3ScoreRequest {
  studentId?: number
  studentName?: string
  examinerName?: string
  remark?: string
  itemRecordValues?: Record<number, Record<string, unknown>>
  itemRecordValueList?: PEP3ItemRecordValueInput[]
}

export interface PEP3AssessmentRecordQueryModel {
  assessmentCode?: string
  studentId?: number
  searchKey?: string
  assessmentDateBegin?: PEP3DateString
  assessmentDateEnd?: PEP3DateString
}

export interface PEP3AssessmentDraftQueryModel extends PEP3AssessmentRecordQueryModel {
  status?: 'draft' | 'ready_to_score' | 'complete' | 'submitted' | string
}

export interface PEP3RecordPageRequest {
  pageRequestModel: PageRequestModel
  queryModel: PEP3AssessmentRecordQueryModel
}

export interface PEP3DraftPageRequest {
  pageRequestModel: PageRequestModel
  queryModel: PEP3AssessmentDraftQueryModel
}

export interface PEP3ScoreOption {
  value: number
  label: string
  description?: string
}

export interface PEP3ItemRecordFieldOption {
  value: string
  label: string
}

export interface PEP3CaregiverReportOption {
  value: string
  label: string
  score?: number
}

export interface PEP3CaregiverReportItem {
  itemNo: number
  key: string
  prompt: string
  fieldType: 'number' | 'radio' | 'textarea' | string
  unit?: string
  required?: boolean
  scored: boolean
  options?: PEP3CaregiverReportOption[]
}

export interface PEP3CaregiverDiagnosisCategory {
  key: string
  label: string
}

export interface PEP3CaregiverReportSection {
  sectionCode: string
  title: string
  description?: string
  inputType: 'age_estimate' | 'diagnosis_matrix' | 'single_choice' | string
  scaleCode?: PEP3ScaleCode
  scaleName?: string
  scored: boolean
  maxRawScore?: number
  items?: PEP3CaregiverReportItem[]
  diagnosisCategories?: PEP3CaregiverDiagnosisCategory[]
}

export interface PEP3CaregiverScoreRule {
  scaleCode: PEP3ScaleCode
  scaleName: string
  sectionCode: string
  maxRawScore: number
  description: string
}

export interface PEP3CaregiverReportTemplate {
  reportName: string
  sourcePdf: string
  submitMode: 'caregiver_self_report' | string
  instructions: string
  scoreRules: PEP3CaregiverScoreRule[]
  sections: PEP3CaregiverReportSection[]
}

export interface PEP3ItemRecordField {
  key: string
  label: string
  fieldType: 'text' | 'textarea' | 'number' | 'radio' | 'checkbox_group' | string
  displayType?: string
  required?: boolean
  placeholder?: string
  options?: PEP3ItemRecordFieldOption[]
}

export interface PEP3AssessmentFormField {
  key: string
  label: string
  fieldType: string
  required: boolean
  placeholder?: string
}

export interface PEP3AssessmentDomain {
  scaleCode: PEP3ScaleCode
  scaleName: string
  category: 'development' | 'behavior' | 'caregiver_report' | string
  itemCount?: number
  maxRawScore?: number
  itemNumbers?: number[]
  isDevelopmentSubtest?: boolean
  isBehaviorSubtest?: boolean
  isCaregiverReport?: boolean
  compositeCode?: string
}

export interface PEP3RawScoreField {
  scaleCode: PEP3ScaleCode
  scaleName: string
  category: string
  minScore: number
  maxScore?: number
  inputMode: 'auto_sum_from_item_scores' | 'manual_raw_score' | string
  required: boolean
  description?: string
}

export interface PEP3AssessmentItem {
  itemNo: number
  itemTitle: string
  testItem: string
  materials?: string
  method?: string
  domainCode: PEP3ScaleCode
  domainName: string
  standard: string
  scoreOptions: PEP3ScoreOption[]
  recordFields?: PEP3ItemRecordField[]
  sourcePdf?: string
  sourcePages?: number[]
  ocrStatus?: string
}

export interface PEP3AssessmentItemGroup {
  groupCode: string
  title: string
  bookletPageNo: number
  sourcePdfPageNo?: number
  layout?: string
  startItemNo: number
  endItemNo: number
  items: PEP3AssessmentItem[]
}

export interface PEP3SubmitContract {
  scoreEndpoint: string
  createRecordEndpoint: string
  dateFormat: string
  itemScoreListKey: string
  rawScoreListKey: string
  itemRecordValuesKey?: string
  itemRecordValueListKey?: string
  requiredBaseFields: string[]
  allowedItemScores: number[]
}

export interface PEP3NormDataInfo {
  normVersion: string
  developmentAgeMaxMonths: number
  normAgeBandMaxMonths: number
  normSourcePdf: string
}

export interface PEP3AssessmentFormTemplate extends PEP3NormDataInfo {
  templateCode: 'PEP3_ASSESSMENT_FORM' | string
  templateVersion: string
  title: string
  scaleCode: 'PEP3' | string
  scaleVersion: string
  dataStatus?: string
  sources?: string[]
  itemCount: number
  scoreOptions: PEP3ScoreOption[]
  basicFields: PEP3AssessmentFormField[]
  domains: PEP3AssessmentDomain[]
  rawScoreFields: PEP3RawScoreField[]
  itemGroups: PEP3AssessmentItemGroup[]
  caregiverReport: PEP3CaregiverReportTemplate
  submitContract: PEP3SubmitContract
}

export interface PEP3AgeResult {
  years: number
  months: number
  days: number
  total_months_for_norm: number
}

export interface PEP3NormValue {
  text: string
  comparator?: string
  number?: number
  table_no?: string
  appendix?: string
  source_pdf?: string
  source_pages?: number[]
  ocr_status?: string
}

export interface PEP3ScaleResult {
  scale_code: PEP3ScaleCode
  scale_name?: string
  raw_score: number
  max_raw_score?: number
  answered_items?: number
  missing_items?: number[]
  development_age?: PEP3NormValue
  percentile_rank?: PEP3NormValue
  scaled_score?: PEP3NormValue
  level?: string
  warnings?: string[]
}

export interface PEP3CompositeMemberScaleScore {
  scale_code: PEP3ScaleCode
  scaled_score?: PEP3NormValue
}

export interface PEP3CompositeResult {
  composite_code: string
  composite_name: string
  member_scale_codes: PEP3ScaleCode[]
  member_scale_scores?: PEP3CompositeMemberScaleScore[]
  standard_score_sum?: number
  percentile_rank?: PEP3NormValue
  level?: string
  development_age_months?: number
  warnings?: string[]
}

export interface PEP3AssessmentResult {
  age: PEP3AgeResult
  scales: Record<PEP3ScaleCode, PEP3ScaleResult>
  composites: Record<string, PEP3CompositeResult>
  warnings?: string[]
}

export interface PEP3ScoreResponse extends PEP3NormDataInfo {
  scaleCode: 'PEP3' | string
  scaleVersion: string
  dataStatus: string
  sources: string[]
  result: PEP3AssessmentResult
}

export interface PEP3AssessmentRecordSummary {
  id: number
  instId: number
  studentId?: number
  studentName?: string
  assessmentCode: string
  assessmentName: string
  scaleVersion: string
  birthDate?: string
  assessmentDate?: string
  ageYears: number
  ageMonths: number
  ageDays: number
  normAgeMonths: number
  examinerId?: number
  examinerName?: string
  dataStatus?: string
  remark?: string
  createdTime?: string
  updatedTime?: string
}

export interface PEP3AssessmentRecordDetail extends PEP3AssessmentRecordSummary {
  input?: Record<string, unknown>
  result?: PEP3ScoreResponse
}

export interface PEP3DomainProgress {
  scaleCode: PEP3ScaleCode
  scaleName: string
  category: string
  itemCount: number
  answeredItemCount: number
  rawScore?: number
  maxRawScore?: number
  complete: boolean
}

export interface PEP3AssessmentDraftProgress {
  itemCount: number
  answeredItemCount: number
  missingItemCount: number
  rawScoreCount: number
  caregiverRawScoreCount: number
  totalInputCount: number
  completedInputCount: number
  completionPercent: number
  complete: boolean
  canScore: boolean
  missingRequiredFields?: string[]
  missingItemNos?: number[]
  domainProgress?: PEP3DomainProgress[]
}

export interface PEP3AssessmentDraftSummary {
  id: number
  instId: number
  studentId?: number
  studentName?: string
  assessmentCode: string
  assessmentName: string
  scaleVersion: string
  birthDate?: string
  assessmentDate?: string
  examinerId?: number
  examinerName?: string
  status: 'draft' | 'ready_to_score' | 'complete' | 'submitted' | string
  submittedRecordId?: number
  answeredItemCount: number
  rawScoreCount: number
  completionPercent: number
  progress: PEP3AssessmentDraftProgress
  remark?: string
  createdTime?: string
  updatedTime?: string
}

export interface PEP3AssessmentDraftDetail extends PEP3AssessmentDraftSummary {
  input?: PEP3DraftSaveRequest
}

export interface PEP3AssessmentDraftSubmitResult {
  draftId: number
  recordId: number
  draftStatus: string
  record: PEP3AssessmentRecordDetail
}

export interface PEP3TemplateField {
  key: string
  label: string
  value: string
  rawValue?: unknown
  unit?: string
  placeholder?: string
}

export interface PEP3TemplateColumn {
  key: string
  label: string
  width?: number
  align?: string
  group?: string
}

export interface PEP3TemplateTable {
  columns: PEP3TemplateColumn[]
  rows: Array<Record<string, unknown>>
  footerRows?: Array<Record<string, unknown>>
}

export interface PEP3TemplateSection {
  sectionCode: string
  title: string
  type: string
  layout?: string
  sourcePdfPageNo?: number
  bookletPageNo?: number
  fields?: PEP3TemplateField[]
  table?: PEP3TemplateTable
  textItems?: string[]
  meta?: Record<string, unknown>
}

export interface PEP3Report extends PEP3NormDataInfo {
  record: PEP3AssessmentRecordSummary
  templateCode: 'PEP3_EXPLANATORY_REPORT' | string
  templateVersion: string
  title: string
  scaleCode: 'PEP3' | string
  scaleVersion: string
  dataStatus?: string
  sources?: string[]
  sections: PEP3TemplateSection[]
}

export interface PEP3BookletPage {
  pageNo: number
  sourcePdfPageNo: number
  title: string
  pageType: string
  sections: PEP3TemplateSection[]
  meta?: Record<string, unknown>
}

export interface PEP3Booklet extends PEP3NormDataInfo {
  record: PEP3AssessmentRecordSummary
  templateCode: 'PEP3_RECORD_BOOKLET' | string
  templateVersion: string
  title: string
  scaleCode: 'PEP3' | string
  scaleVersion: string
  dataStatus?: string
  sources?: string[]
  sourcePdf: string
  pages: PEP3BookletPage[]
  warnings?: string[]
}

export function getPEP3AssessmentFormTemplateApi() {
  return useGet<PEP3AssessmentFormTemplate>('/api/v1/assessments/pep3/form-template')
}

export function scorePEP3AssessmentApi(data: PEP3ScoreRequest) {
  return usePost<PEP3ScoreResponse>('/api/v1/assessments/pep3/score', data)
}

export function savePEP3AssessmentDraftApi(data: PEP3DraftSaveRequest) {
  return usePost<PEP3AssessmentDraftDetail>('/api/v1/assessments/pep3/drafts/save', data)
}

export function getPEP3AssessmentDraftDetailApi(id: number) {
  return useGet<PEP3AssessmentDraftDetail>('/api/v1/assessments/pep3/drafts/detail', { id })
}

export function pagePEP3AssessmentDraftsApi(data: PEP3DraftPageRequest) {
  return usePost<PageResult<PEP3AssessmentDraftSummary>>('/api/v1/assessments/pep3/drafts/page', data)
}

export function submitPEP3AssessmentDraftApi(id: number) {
  return usePost<PEP3AssessmentDraftSubmitResult>('/api/v1/assessments/pep3/drafts/submit', { id })
}

export function deletePEP3AssessmentDraftApi(id: number) {
  return usePost<boolean>('/api/v1/assessments/pep3/drafts/delete', { id })
}

export function createPEP3AssessmentRecordApi(data: PEP3RecordCreateRequest) {
  return usePost<PEP3AssessmentRecordDetail>('/api/v1/assessments/pep3/records/create', data)
}

export function getPEP3AssessmentRecordDetailApi(id: number) {
  return useGet<PEP3AssessmentRecordDetail>('/api/v1/assessments/pep3/records/detail', { id })
}

export function getPEP3AssessmentReportApi(id: number) {
  return useGet<PEP3Report>('/api/v1/assessments/pep3/records/report', { id })
}

export function getPEP3AssessmentBookletApi(id: number) {
  return useGet<PEP3Booklet>('/api/v1/assessments/pep3/records/booklet', { id })
}

export function downloadPEP3AssessmentBookletPdfApi(id: number) {
  const token = useAuthorization()
  return axios.get('/api/v1/assessments/pep3/records/booklet/pdf', {
    params: { id },
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export function pagePEP3AssessmentRecordsApi(data: PEP3RecordPageRequest) {
  return usePost<PageResult<PEP3AssessmentRecordSummary>>('/api/v1/assessments/pep3/records/page', data)
}
