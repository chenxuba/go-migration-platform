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
  caregiverReport?: PEP3CaregiverReportSubmission
}

export interface PEP3DraftItemSaveRequest {
  draftId: number
  itemNo: number
  score?: number
  recordValues?: Record<string, unknown>
}

export interface PEP3RecordCreateRequest extends PEP3ScoreRequest {
  studentId?: number
  studentName?: string
  examinerName?: string
  remark?: string
  itemRecordValues?: Record<number, Record<string, unknown>>
  itemRecordValueList?: PEP3ItemRecordValueInput[]
  caregiverReport?: PEP3CaregiverReportSubmission
}

export interface PEP3RecordUpdateRequest extends PEP3RecordCreateRequest {
  id: number
}

export interface PEP3RecordConfigUpdateRequest {
  id: number
  examinerName: string
  assessmentDate: PEP3DateString
}

export interface PEP3AssessmentRecordQueryModel {
  assessmentCode?: string
  scaleCategory?: string
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

function normalizePEP3PageRequest<T extends PEP3RecordPageRequest | PEP3DraftPageRequest>(data: T): T {
  const normalized = {
    ...data,
    queryModel: {
      ...data.queryModel,
    },
  } as T
  const rawStudentId = normalized.queryModel?.studentId
  if (rawStudentId !== undefined && rawStudentId !== null && `${rawStudentId}`.trim() !== '') {
    const studentId = Number(rawStudentId)
    if (Number.isFinite(studentId) && studentId > 0)
      normalized.queryModel.studentId = studentId
  }
  return normalized
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

export interface PEP3CaregiverRawScore {
  scaleCode: PEP3ScaleCode
  rawScore: number
}

export interface PEP3CaregiverReportSubmission {
  respondentName?: string
  relationship?: string
  answers?: Record<string, Record<string, unknown>>
  rawScores?: Partial<Record<PEP3ScaleCode, number>>
  rawScoreList?: PEP3CaregiverRawScore[]
  submittedAt?: string
  source?: string
}

export interface PEP3CaregiverReportInvite {
  draftId: number
  recordId?: number
  studentName?: string
  ticket?: string
  token: string
  expiresAt?: string
  miniProgramPath: string
  miniProgramEnvVersion?: 'develop' | 'trial' | 'release' | string
  miniProgramCodeDataUrl?: string
  wechatUrlLink?: string
  qrCodeValue?: string
  qrCodeType?: 'wechat_mini_program_code' | 'wechat_url_link' | 'mini_program_path' | string
  qrCodeMessage?: string
  url: string
}

export interface PEP3CaregiverReportPublicTemplate {
  draftId: number
  studentId?: number
  studentName?: string
  birthDate?: string
  assessmentDate?: string
  template: PEP3CaregiverReportTemplate
  submission?: PEP3CaregiverReportSubmission
}

export interface PEP3CaregiverReportSubmitRequest {
  token: string
  respondentName?: string
  relationship?: string
  answers: Record<string, Record<string, unknown>>
}

export interface PEP3CaregiverReportSubmitResult {
  draftId: number
  recordId?: number
  recordUpdated?: boolean
  studentName?: string
  rawScores: Partial<Record<PEP3ScaleCode, number>>
  progress: PEP3AssessmentDraftProgress
  submittedAt?: string
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
  materialImages?: string[]
  method?: string
  guidance?: string
  guidanceVideo?: string
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

export interface PEP3AssessmentItemSummary {
  itemNo: number
  itemTitle: string
  testItem: string
  domainCode: PEP3ScaleCode
  domainName: string
}

export interface PEP3AssessmentItemGroupSummary {
  groupCode: string
  title: string
  bookletPageNo: number
  sourcePdfPageNo?: number
  layout?: string
  startItemNo: number
  endItemNo: number
  items: PEP3AssessmentItemSummary[]
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

export interface PEP3AssessmentFormTemplateSummary extends PEP3NormDataInfo {
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
  itemGroups: PEP3AssessmentItemGroupSummary[]
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
  studentGender?: string
  studentAvatar?: string
  assessmentCode: string
  assessmentName: string
  scaleCategory?: string
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
  iepPlanStatus?: string
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

export interface PEP3ReportInterpretation {
  title: string
  model?: string
  generatedBy: string
  generatedAt?: string
  summary: string
  domainAnalysis: string[]
  suggestions: string[]
  notes?: string[]
}

export interface PEP3ReportInterpretationStreamHandlers {
  onStatus?: (message: string) => void
  onDelta?: (text: string) => void
  onDone?: (data: PEP3ReportInterpretation) => void
}

export interface PEP3ReportInterpretationStreamOptions {
  signal?: AbortSignal
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

export function getPEP3AssessmentFormTemplateSummaryApi() {
  return useGet<PEP3AssessmentFormTemplateSummary>('/api/v1/assessments/pep3/form-template/summary')
}

export function getPEP3AssessmentFormTemplateItemApi(itemNo: number) {
  return useGet<PEP3AssessmentItem>('/api/v1/assessments/pep3/form-template/item', { itemNo })
}

export function scorePEP3AssessmentApi(data: PEP3ScoreRequest) {
  return usePost<PEP3ScoreResponse>('/api/v1/assessments/pep3/score', data)
}

export function savePEP3AssessmentDraftApi(data: PEP3DraftSaveRequest) {
  return usePost<PEP3AssessmentDraftDetail>('/api/v1/assessments/pep3/drafts/save', data)
}

export function savePEP3AssessmentDraftItemApi(data: PEP3DraftItemSaveRequest) {
  return usePost<PEP3AssessmentDraftDetail>('/api/v1/assessments/pep3/drafts/item/save', data)
}

export function getPEP3AssessmentDraftDetailApi(id: number) {
  return useGet<PEP3AssessmentDraftDetail>('/api/v1/assessments/pep3/drafts/detail', { id })
}

export function pagePEP3AssessmentDraftsApi(data: PEP3DraftPageRequest) {
  return usePost<PageResult<PEP3AssessmentDraftSummary>>('/api/v1/assessments/pep3/drafts/page', normalizePEP3PageRequest(data), { silentError: true })
}

export function invitePEP3CaregiverReportApi(draftId: number) {
  return usePost<PEP3CaregiverReportInvite>('/api/v1/assessments/pep3/drafts/caregiver-report/invite', { draftId })
}

export function invitePEP3CaregiverReportForRecordApi(recordId: number) {
  return usePost<PEP3CaregiverReportInvite>('/api/v1/assessments/pep3/records/caregiver-report/invite', { recordId })
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

export function updatePEP3AssessmentRecordApi(data: PEP3RecordUpdateRequest) {
  return usePost<PEP3AssessmentRecordDetail>('/api/v1/assessments/pep3/records/update', data)
}

export function updatePEP3AssessmentRecordConfigApi(data: PEP3RecordConfigUpdateRequest) {
  return usePost<PEP3AssessmentRecordDetail>('/api/v1/assessments/pep3/records/config/update', data)
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

export type PEP3BookletPdfExportDimension =
  | 'test_score'
  | 'development_profile'
  | 'score_and_profile'
  | 'scoring_tables'
  | 'education_plan'
  | 'all'

export function downloadPEP3AssessmentBookletPdfApi(id: number, dimension: PEP3BookletPdfExportDimension = 'all') {
  const token = useAuthorization()
  return axios.get('/api/v1/assessments/pep3/records/booklet/pdf', {
    params: { id, dimension },
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export function getPEP3AssessmentRecordReportInterpretationApi(id: number) {
  return useGet<PEP3ReportInterpretation>('/api/v1/assessments/pep3/records/report/interpretation', { id }, { silentError: true })
}

export function generatePEP3AssessmentRecordReportInterpretationApi(id: number) {
  return usePost<PEP3ReportInterpretation>('/api/v1/assessments/pep3/records/report/interpretation/ai', { id }, {
    loading: false,
    silentError: true,
    timeout: 190000,
  })
}

function pep3StreamHeaders(accept = 'text/event-stream') {
  const token = useAuthorization()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    [STORAGE_AUTHORIZE_KEY]: token.value || '',
    Authorization: token.value ? `Bearer ${token.value}` : '',
    'Accept-Language': 'zh-CN',
    Accept: accept,
  }
  if (typeof window !== 'undefined')
    headers['X-Tenant-Domain'] = window.location.hostname.toLowerCase()
  return headers
}

export async function generatePEP3AssessmentRecordReportInterpretationStreamApi(
  id: number,
  handlers: PEP3ReportInterpretationStreamHandlers = {},
  options: PEP3ReportInterpretationStreamOptions = {},
) {
  const response = await fetch('/api/v1/assessments/pep3/records/report/interpretation/ai/stream', {
    method: 'POST',
    headers: pep3StreamHeaders(),
    body: JSON.stringify({ id }),
    signal: options.signal,
  })
  if (!response.ok) {
    const text = await response.text()
    let message = text || '报告解读生成失败'
    try {
      const payload = JSON.parse(text)
      message = payload?.message || message
    }
    catch {
    }
    throw new Error(message)
  }
  if (!response.body)
    throw new Error('当前浏览器不支持流式生成')

  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''
  let finalInterpretation: PEP3ReportInterpretation | null = null

  function handleFrame(frame: string) {
    const lines = frame.split('\n')
    const dataLines = lines
      .filter(line => line.startsWith('data:'))
      .map(line => line.slice(5).trim())
    if (!dataLines.length)
      return
    const payload = JSON.parse(dataLines.join('\n'))
    if (payload?.type === 'status')
      handlers.onStatus?.(payload.message || '')
    else if (payload?.type === 'delta')
      handlers.onDelta?.(payload.text || '')
    else if (payload?.type === 'done') {
      finalInterpretation = payload.data
      if (finalInterpretation)
        handlers.onDone?.(finalInterpretation)
    }
    else if (payload?.type === 'error') {
      throw new Error(payload.message || '报告解读生成失败')
    }
  }

  while (true) {
    const { value, done } = await reader.read()
    buffer += decoder.decode(value || new Uint8Array(), { stream: !done })
    const frames = buffer.split(/\n\n/)
    buffer = frames.pop() || ''
    for (const frame of frames) {
      if (frame.trim())
        handleFrame(frame)
    }
    if (done)
      break
  }
  if (buffer.trim())
    handleFrame(buffer)
  if (options.signal?.aborted)
    throw new DOMException('报告解读生成已取消', 'AbortError')
  if (!finalInterpretation)
    throw new Error('报告解读生成未返回结果')
  return finalInterpretation
}

export function downloadPEP3IEPPlanWordApi(params: { id?: number | string, duration?: number | string, plan?: PEP3IEPPlanAIResult } = {}) {
  const token = useAuthorization()
  const headers = {
    [STORAGE_AUTHORIZE_KEY]: token.value || '',
    Authorization: token.value ? `Bearer ${token.value}` : '',
    'Accept-Language': 'zh-CN',
  }
  if (params.plan) {
    return axios.post('/api/v1/assessments/pep3/records/iep-plan/word', {
      id: Number(params.id || 0),
      durationMonths: Number(params.duration || 0),
      plan: params.plan,
    }, {
      responseType: 'blob',
      headers: {
        ...headers,
        'Content-Type': 'application/json',
      },
    })
  }
  return axios.get('/api/v1/assessments/pep3/records/iep-plan/word', {
    params,
    responseType: 'blob',
    headers,
  })
}

export interface PEP3IEPPlanAIResult {
  title: string
  model?: string
  student: {
    name: string
    gender: string
    birthDate: string
  }
  meta: {
    planDate: string
    participant: string
    implementer: string
    startDate: string
    endDate: string
  }
  rows: Array<{
    domain: string
    longGoal: string
    shortGoal: string
    courseForm: string
    startEndDate: string
  }>
}

export interface PEP3MonthlyPlanAIResult {
  title: string
  model?: string
  student: PEP3IEPPlanAIResult['student']
  meta: {
    planDate: string
    participant: string
    implementer: string
    startDate: string
    endDate: string
    monthLabel?: string
    sourceTitle?: string
  }
  rows: Array<{
    domain: string
    longGoal: string
    shortGoal: string
    trainingItems: Array<{
      content: string
      startEndDate: string
    }>
    courseForm: string
  }>
}

export interface PEP3WeeklyPlanAIResult {
  title: string
  model?: string
  student: PEP3IEPPlanAIResult['student']
  teacherName: string
  courseName: string
  trainingDate: string
  preparation: string
  weekDates: string[]
  rows: Array<{
    project: string
    content: string
    completion?: string[]
  }>
  sourceTitle?: string
}

export interface PEP3IEPPlanSavedVO {
  exists: boolean
  status?: 'draft' | 'confirmed' | string
  durationMonths?: number
  plan?: PEP3IEPPlanAIResult
  updatedTime?: string
}

export interface PEP3ExecutionPlanSavedVO {
  exists: boolean
  durationMonths?: number
  monthlyPlans?: Array<{
    targetMonthIndex: number
    plan: PEP3MonthlyPlanAIResult
    updatedTime?: string
  }>
  weeklyPlans?: Array<{
    targetMonthIndex: number
    targetWeekIndex: number
    plan: PEP3WeeklyPlanAIResult
    updatedTime?: string
  }>
}

export function generatePEP3IEPPlanAIApi(data: { id?: number | string, durationMonths?: number | string }) {
  return usePost<PEP3IEPPlanAIResult>('/api/v1/assessments/pep3/records/iep-plan/ai', data, {
    loading: false,
    silentError: true,
    timeout: 180000,
  })
}

export function generatePEP3ExecutionPlanAIApi(data: {
  id?: number | string
  durationMonths?: number | string
  planType: 'monthly' | 'weekly'
  targetMonthIndex?: number | string
  targetWeekIndex?: number | string
  sourcePlan: PEP3IEPPlanAIResult
  monthlyPlan?: PEP3MonthlyPlanAIResult | null
}) {
  return usePost<PEP3MonthlyPlanAIResult | PEP3WeeklyPlanAIResult>('/api/v1/assessments/pep3/records/iep-plan/execution/ai', {
    ...data,
    id: Number(data.id || 0),
    durationMonths: Number(data.durationMonths || 0),
    targetMonthIndex: Number(data.targetMonthIndex || 0),
    targetWeekIndex: Number(data.targetWeekIndex || 0),
  }, {
    loading: false,
    silentError: true,
    timeout: 180000,
  })
}

export interface PEP3ExecutionPlanAIStreamHandlers {
  onStatus?: (message: string) => void
  onDelta?: (text: string) => void
  onDone?: (data: PEP3MonthlyPlanAIResult | PEP3WeeklyPlanAIResult) => void
}

export async function generatePEP3ExecutionPlanAIStreamApi(
  data: {
    id?: number | string
    durationMonths?: number | string
    planType: 'monthly' | 'weekly'
    targetMonthIndex?: number | string
    targetWeekIndex?: number | string
    sourcePlan: PEP3IEPPlanAIResult
    monthlyPlan?: PEP3MonthlyPlanAIResult | null
  },
  handlers: PEP3ExecutionPlanAIStreamHandlers = {},
  options: PEP3IEPPlanAIStreamOptions = {},
) {
  const token = useAuthorization()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    [STORAGE_AUTHORIZE_KEY]: token.value || '',
    Authorization: token.value ? `Bearer ${token.value}` : '',
    'Accept-Language': 'zh-CN',
  }
  if (typeof window !== 'undefined')
    headers['X-Tenant-Domain'] = window.location.hostname.toLowerCase()

  const response = await fetch('/api/v1/assessments/pep3/records/iep-plan/execution/ai/stream', {
    method: 'POST',
    headers,
    body: JSON.stringify({
      ...data,
      id: Number(data.id || 0),
      durationMonths: Number(data.durationMonths || 0),
      targetMonthIndex: Number(data.targetMonthIndex || 0),
      targetWeekIndex: Number(data.targetWeekIndex || 0),
    }),
    signal: options.signal,
  })
  if (!response.ok) {
    const text = await response.text()
    let message = text || 'AI生成失败'
    try {
      const payload = JSON.parse(text)
      message = payload?.message || message
    }
    catch {
    }
    throw new Error(message)
  }
  if (!response.body)
    throw new Error('当前浏览器不支持流式生成')

  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''
  let finalPlan: PEP3MonthlyPlanAIResult | PEP3WeeklyPlanAIResult | null = null

  function handleFrame(frame: string) {
    const lines = frame.split('\n')
    const dataLines = lines
      .filter(line => line.startsWith('data:'))
      .map(line => line.slice(5).trim())
    if (!dataLines.length)
      return
    const payload = JSON.parse(dataLines.join('\n'))
    if (payload?.type === 'status')
      handlers.onStatus?.(payload.message || '')
    else if (payload?.type === 'delta')
      handlers.onDelta?.(payload.text || '')
    else if (payload?.type === 'done') {
      finalPlan = payload.data
      if (finalPlan)
        handlers.onDone?.(finalPlan)
    }
    else if (payload?.type === 'error') {
      throw new Error(payload.message || 'AI生成失败')
    }
  }

  while (true) {
    const { value, done } = await reader.read()
    buffer += decoder.decode(value || new Uint8Array(), { stream: !done })
    const frames = buffer.split(/\n\n/)
    buffer = frames.pop() || ''
    for (const frame of frames) {
      if (frame.trim())
        handleFrame(frame)
    }
    if (done)
      break
  }
  if (buffer.trim())
    handleFrame(buffer)
  if (options.signal?.aborted)
    throw new DOMException('AI生成已取消', 'AbortError')
  if (!finalPlan)
    throw new Error('AI生成未返回计划数据')
  return finalPlan
}

export function downloadPEP3ExecutionPlanWordApi(data: {
  id?: number | string
  planType: 'monthly' | 'weekly'
  monthlyPlan?: PEP3MonthlyPlanAIResult | null
  weeklyPlan?: PEP3WeeklyPlanAIResult | null
}) {
  const token = useAuthorization()
  return axios.post('/api/v1/assessments/pep3/records/iep-plan/execution/word', {
    ...data,
    id: Number(data.id || 0),
  }, {
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
      'Content-Type': 'application/json',
    },
  })
}

export function getPEP3ExecutionPlansApi(id: number | string, durationMonths?: number | string) {
  return useGet<PEP3ExecutionPlanSavedVO>('/api/v1/assessments/pep3/records/iep-plan/execution/detail', {
    id,
    durationMonths: Number(durationMonths || 0),
  }, {
    loading: false,
    silentError: true,
  })
}

export function savePEP3ExecutionPlanApi(data: {
  id?: number | string
  durationMonths?: number | string
  planType: 'monthly' | 'weekly'
  targetMonthIndex?: number | string
  targetWeekIndex?: number | string
  monthlyPlan?: PEP3MonthlyPlanAIResult | null
  weeklyPlan?: PEP3WeeklyPlanAIResult | null
  preserveWeeklyPlans?: boolean
}) {
  return usePost<PEP3ExecutionPlanSavedVO>('/api/v1/assessments/pep3/records/iep-plan/execution/save', {
    ...data,
    id: Number(data.id || 0),
    durationMonths: Number(data.durationMonths || 0),
    targetMonthIndex: Number(data.targetMonthIndex || 0),
    targetWeekIndex: Number(data.targetWeekIndex || 0),
  }, {
    loading: false,
    silentError: true,
    timeout: 60000,
  })
}

export function getPEP3IEPPlanApi(id: number | string, durationMonths?: number | string) {
  return useGet<PEP3IEPPlanSavedVO>('/api/v1/assessments/pep3/records/iep-plan/detail', {
    id,
    durationMonths: Number(durationMonths || 0),
  }, {
    loading: false,
    silentError: true,
  })
}

export function savePEP3IEPPlanApi(data: {
  id?: number | string
  durationMonths?: number | string
  status?: string
  plan: PEP3IEPPlanAIResult
}) {
  return usePost<PEP3IEPPlanSavedVO>('/api/v1/assessments/pep3/records/iep-plan/save', {
    ...data,
    id: Number(data.id || 0),
    durationMonths: Number(data.durationMonths || 0),
  }, {
    loading: false,
    silentError: true,
    timeout: 60000,
  })
}

export interface PEP3IEPPlanAIStreamHandlers {
  onStatus?: (message: string) => void
  onDelta?: (text: string) => void
  onDone?: (data: PEP3IEPPlanAIResult) => void
}

export interface PEP3IEPPlanAIStreamOptions {
  signal?: AbortSignal
}

export async function generatePEP3IEPPlanAIStreamApi(
  data: { id?: number | string, durationMonths?: number | string },
  handlers: PEP3IEPPlanAIStreamHandlers = {},
  options: PEP3IEPPlanAIStreamOptions = {},
) {
  const token = useAuthorization()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    [STORAGE_AUTHORIZE_KEY]: token.value || '',
    Authorization: token.value ? `Bearer ${token.value}` : '',
    'Accept-Language': 'zh-CN',
  }
  if (typeof window !== 'undefined')
    headers['X-Tenant-Domain'] = window.location.hostname.toLowerCase()

  const response = await fetch('/api/v1/assessments/pep3/records/iep-plan/ai/stream', {
    method: 'POST',
    headers,
    body: JSON.stringify(data),
    signal: options.signal,
  })
  if (!response.ok) {
    const text = await response.text()
    let message = text || 'AI生成失败'
    try {
      const payload = JSON.parse(text)
      message = payload?.message || message
    }
    catch {
    }
    throw new Error(message)
  }
  if (!response.body)
    throw new Error('当前浏览器不支持流式生成')

  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''
  let finalPlan: PEP3IEPPlanAIResult | null = null

  function handleFrame(frame: string) {
    const lines = frame.split('\n')
    const dataLines = lines
      .filter(line => line.startsWith('data:'))
      .map(line => line.slice(5).trim())
    if (!dataLines.length)
      return
    const payload = JSON.parse(dataLines.join('\n'))
    if (payload?.type === 'status')
      handlers.onStatus?.(payload.message || '')
    else if (payload?.type === 'delta')
      handlers.onDelta?.(payload.text || '')
    else if (payload?.type === 'done') {
      finalPlan = payload.data
      if (finalPlan)
        handlers.onDone?.(finalPlan)
    }
    else if (payload?.type === 'error') {
      throw new Error(payload.message || 'AI生成失败')
    }
  }

  while (true) {
    const { value, done } = await reader.read()
    buffer += decoder.decode(value || new Uint8Array(), { stream: !done })
    const frames = buffer.split(/\n\n/)
    buffer = frames.pop() || ''
    for (const frame of frames) {
      if (frame.trim())
        handleFrame(frame)
    }
    if (done)
      break
  }
  if (buffer.trim())
    handleFrame(buffer)
  if (options.signal?.aborted)
    throw new DOMException('AI生成已取消', 'AbortError')
  if (!finalPlan)
    throw new Error('AI生成未返回计划数据')
  return finalPlan
}

export function pagePEP3AssessmentRecordsApi(data: PEP3RecordPageRequest) {
  return usePost<PageResult<PEP3AssessmentRecordSummary>>('/api/v1/assessments/pep3/records/page', normalizePEP3PageRequest(data), { silentError: true })
}

export function deletePEP3AssessmentRecordApi(id: number) {
  return usePost<boolean>('/api/v1/assessments/pep3/records/delete', { id })
}
