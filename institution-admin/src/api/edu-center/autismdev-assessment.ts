import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { useGet, usePost } from '~/utils/request'
import type {
  PageRequestModel,
  PageResult,
  PEP3ExecutionPlanSavedVO,
  PEP3IEPPlanAIResult,
  PEP3IEPPlanAIStreamOptions,
  PEP3IEPPlanPeriodSyncVO,
  PEP3IEPPlanSavedVO,
  PEP3AssessmentRecordDetail,
  PEP3AssessmentRecordQueryModel,
  PEP3AssessmentRecordSummary,
  PEP3DomainProgress,
  PEP3MonthlyPlanAIResult,
  PEP3WeeklyPlanAIResult,
} from './pep3-assessment'

export type AutismDevScaleCode = 'AUTISMDEV' | string
export type AutismDevQuestionDisplayPreference = 'all' | 'matchingAge' | 'ageAndBelow' | string
export type AutismDevScopeMode = 'full' | 'custom' | string

export interface AutismDevScoreOption {
  value: string
  label: string
  description?: string
  scoreType: string
}

export interface AutismDevDomain {
  domainCode: string
  domainName: string
  sortNo: number
  itemCount: number
  scoreType: string
}

export interface AutismDevItemSummary {
  itemNo: number
  domainItemNo: number
  itemTitle: string
  testItem: string
  assessmentRange: string
  materials: string
  method: string
  passCriteria: string
  ageSegment: string
  ageMinMonth: number
  ageMaxMonth: number
  domainCode: string
  domainName: string
  scoreType: string
  assessmentModes: string[]
}

export interface AutismDevAssessmentItem extends AutismDevItemSummary {
  scoreOptions: AutismDevScoreOption[]
}

export interface AutismDevDomainGroup {
  groupCode: string
  title: string
  domainCode: string
  domainName: string
  scoreType: string
  itemCount: number
  items: AutismDevItemSummary[]
}

export interface AutismDevTemplateSummary {
  templateCode: string
  title: string
  scaleCode: AutismDevScaleCode
  scaleVersion: string
  itemCount: number
  scoreOptions: AutismDevScoreOption[]
  domains: AutismDevDomain[]
  domainGroups: AutismDevDomainGroup[]
}

export interface AutismDevItemScoreInput {
  itemNo: number
  score: string
  remark?: string
}

export interface AutismDevItemRemarkInput {
  itemNo: number
  remark: string
}

export interface AutismDevDraftInput {
  studentId?: number
  studentName?: string
  examinerName?: string
  remark?: string
  birthDate?: string
  assessmentDate?: string
  scopeMode?: AutismDevScopeMode
  scopeDomainCodes?: string[]
  questionDisplayPreference?: AutismDevQuestionDisplayPreference
  itemScores?: Record<number, string>
  itemScoreList?: AutismDevItemScoreInput[]
  itemRemarks?: Record<number, string>
  itemRemarkList?: AutismDevItemRemarkInput[]
}

export interface AutismDevDraftProgress {
  itemCount: number
  answeredItemCount: number
  missingItemCount: number
  rawScoreCount?: number
  totalInputCount?: number
  completedInputCount?: number
  completionPercent: number
  complete: boolean
  canScore: boolean
  questionDisplayPreference?: AutismDevQuestionDisplayPreference
  missingRequiredFields?: string[]
  missingItemNos?: number[]
  domainProgress?: PEP3DomainProgress[]
}

export interface AutismDevAssessmentDraftSummary {
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
  progress: AutismDevDraftProgress
  remark?: string
  createdTime?: string
  updatedTime?: string
}

export interface AutismDevDraftDetail extends AutismDevAssessmentDraftSummary {
  input?: AutismDevDraftInput
}

export interface AutismDevAssessmentDraftQueryModel extends PEP3AssessmentRecordQueryModel {
  status?: 'draft' | 'ready_to_score' | 'complete' | 'submitted' | string
  latestOnly?: boolean
}

export interface AutismDevDraftPageRequest {
  pageRequestModel: PageRequestModel
  queryModel: AutismDevAssessmentDraftQueryModel
  latestOnly?: boolean
}

export interface AutismDevDraftSaveRequest extends AutismDevDraftInput {
  id?: number
}

export interface AutismDevDraftItemSaveRequest {
  draftId: number
  itemNo: number
  score: string
  remark?: string
}

export interface AutismDevAssessmentDraftSubmitResult {
  draftId: number
  recordId: number
  draftStatus: string
  record: PEP3AssessmentRecordSummary
}

export interface AutismDevRecordPageRequest {
  pageRequestModel: PageRequestModel
  queryModel: PEP3AssessmentRecordQueryModel
}

export type AutismDevAssessmentRecordSummary = PEP3AssessmentRecordSummary

export interface AutismDevAssessmentRecordDetail extends AutismDevAssessmentRecordSummary {
  input?: AutismDevDraftInput
  result?: any
}

export interface AutismDevRecordConfigUpdateRequest {
  id: number
  examinerName: string
  assessmentDate: string
}

export type AutismDevSelectedReportSection =
  | 'assessmentInfo'
  | 'resultAnalysis'
  | 'training'
  | 'developmentProfile'
  | 'behaviorProfile'

export interface AutismDevResultAnalysisRow {
  domain: string
  status: string
  strengths: string
  weaknesses: string
  targets: string
}

export interface AutismDevResultAnalysis {
  title: string
  model?: string
  generatedBy?: string
  generatedAt?: string
  rows: AutismDevResultAnalysisRow[]
}

export interface AutismDevResultAnalysisStreamHandlers {
  onStatus?: (message: string) => void
  onDelta?: (text: string) => void
  onDone?: (data: AutismDevResultAnalysis) => void
}

export interface AutismDevIEPPlanAIStreamHandlers {
  onStatus?: (message: string) => void
  onDelta?: (text: string) => void
  onDone?: (data: PEP3IEPPlanAIResult) => void
}

export interface AutismDevExecutionPlanAIStreamHandlers {
  onStatus?: (message: string) => void
  onDelta?: (text: string) => void
  onDone?: (data: PEP3MonthlyPlanAIResult | PEP3WeeklyPlanAIResult) => void
}

function normalizeAutismDevPageRequest<T extends AutismDevDraftPageRequest>(data: T): T {
  const normalized = {
    ...data,
    queryModel: {
      ...data.queryModel,
      assessmentCode: 'AUTISMDEV',
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

function normalizeAutismDevRecordPageRequest<T extends AutismDevRecordPageRequest>(data: T): T {
  const normalized = {
    ...data,
    queryModel: {
      ...data.queryModel,
      assessmentCode: 'AUTISMDEV',
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

function autismDevAuthHeaders(extra: Record<string, string> = {}) {
  const token = useAuthorization()
  const headers: Record<string, string> = {
    [STORAGE_AUTHORIZE_KEY]: token.value || '',
    Authorization: token.value ? `Bearer ${token.value}` : '',
    'Accept-Language': 'zh-CN',
    ...extra,
  }
  if (typeof window !== 'undefined')
    headers['X-Tenant-Domain'] = window.location.hostname.toLowerCase()
  return headers
}

function autismDevStreamHeaders() {
  return autismDevAuthHeaders({
    'Content-Type': 'application/json',
    Accept: 'text/event-stream',
  })
}

async function readAutismDevSSE<T>(
  response: Response,
  handlers: {
    onStatus?: (message: string) => void
    onDelta?: (text: string) => void
    onDone?: (data: T) => void
  },
  options: PEP3IEPPlanAIStreamOptions,
  fallbackError: string,
  emptyResultError: string,
) {
  if (!response.ok) {
    const text = await response.text()
    let message = text || fallbackError
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
  let finalData: T | null = null

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
      finalData = payload.data
      if (finalData)
        handlers.onDone?.(finalData)
    }
    else if (payload?.type === 'error') {
      throw new Error(payload.message || fallbackError)
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
  if (!finalData)
    throw new Error(emptyResultError)
  return finalData
}

export function getAutismDevAssessmentFormTemplateSummaryApi() {
  return useGet<AutismDevTemplateSummary>('/api/v1/assessments/autismdev/form-template/summary')
}

export function getAutismDevAssessmentFormTemplateItemApi(itemNo: number) {
  return useGet<AutismDevAssessmentItem>('/api/v1/assessments/autismdev/form-template/item', { itemNo })
}

export function saveAutismDevAssessmentDraftApi(data: AutismDevDraftSaveRequest) {
  return usePost<AutismDevDraftDetail>('/api/v1/assessments/autismdev/drafts/save', data)
}

export function saveAutismDevAssessmentDraftItemApi(data: AutismDevDraftItemSaveRequest) {
  return usePost<AutismDevDraftDetail>('/api/v1/assessments/autismdev/drafts/item/save', data)
}

export function getAutismDevAssessmentDraftDetailApi(id: number) {
  return useGet<AutismDevDraftDetail>('/api/v1/assessments/autismdev/drafts/detail', { id })
}

export function pageAutismDevAssessmentDraftsApi(data: AutismDevDraftPageRequest) {
  return usePost<PageResult<AutismDevAssessmentDraftSummary>>(
    '/api/v1/assessments/autismdev/drafts/page',
    normalizeAutismDevPageRequest(data),
    { silentError: true },
  )
}

export function submitAutismDevAssessmentDraftApi(id: number) {
  return usePost<AutismDevAssessmentDraftSubmitResult>('/api/v1/assessments/autismdev/drafts/submit', { id })
}

export function pageAutismDevAssessmentRecordsApi(data: AutismDevRecordPageRequest) {
  return usePost<PageResult<PEP3AssessmentRecordSummary>>(
    '/api/v1/assessments/autismdev/records/page',
    normalizeAutismDevRecordPageRequest(data),
    { silentError: true },
  )
}

export function getAutismDevAssessmentRecordDetailApi(id: number) {
  return useGet<AutismDevAssessmentRecordDetail>('/api/v1/assessments/autismdev/records/detail', { id })
}

export function updateAutismDevAssessmentRecordConfigApi(data: AutismDevRecordConfigUpdateRequest) {
  return usePost<PEP3AssessmentRecordDetail>('/api/v1/assessments/autismdev/records/config/update', data)
}

export function deleteAutismDevAssessmentRecordApi(id: number) {
  return usePost<boolean>('/api/v1/assessments/autismdev/records/delete', { id })
}

export function getAutismDevResultAnalysisApi(id: number | string) {
  return useGet<AutismDevResultAnalysis>('/api/v1/assessments/autismdev/records/result-analysis', { id }, {
    loading: false,
    silentError: true,
  })
}

export function saveAutismDevResultAnalysisApi(id: number | string, analysis: AutismDevResultAnalysis) {
  return usePost<AutismDevResultAnalysis>('/api/v1/assessments/autismdev/records/result-analysis', {
    id: Number(id || 0),
    analysis,
  }, {
    loading: false,
    silentError: true,
    timeout: 60000,
  })
}

export async function generateAutismDevResultAnalysisStreamApi(
  id: number | string,
  handlers: AutismDevResultAnalysisStreamHandlers = {},
  options: PEP3IEPPlanAIStreamOptions = {},
) {
  const response = await fetch('/api/v1/assessments/autismdev/records/result-analysis/ai/stream', {
    method: 'POST',
    headers: autismDevStreamHeaders(),
    body: JSON.stringify({ id: Number(id || 0) }),
    signal: options.signal,
  })
  return readAutismDevSSE<AutismDevResultAnalysis>(
    response,
    handlers,
    options,
    '评估结果分析生成失败',
    '评估结果分析生成未返回结果',
  )
}

export function downloadAutismDevSelectedReportPdfApi(
  id: number,
  sections: AutismDevSelectedReportSection[] = ['assessmentInfo', 'developmentProfile', 'behaviorProfile'],
  analysis?: AutismDevResultAnalysis | null,
) {
  return axios.post('/api/v1/assessments/autismdev/records/selected-report/pdf', {
    id,
    sections,
    ...(analysis ? { analysis } : {}),
  }, {
    responseType: 'blob',
    headers: autismDevAuthHeaders({
      'Accept-Language': 'zh-CN',
      'Content-Type': 'application/json',
    }),
  })
}

export function downloadAutismDevResultAnalysisWordApi(
  id: number | string,
  analysis?: AutismDevResultAnalysis | null,
) {
  if (analysis != null) {
    return axios.post('/api/v1/assessments/autismdev/records/result-analysis/word', {
      id: Number(id || 0),
      analysis,
    }, {
      responseType: 'blob',
      headers: autismDevAuthHeaders({
        'Content-Type': 'application/json',
      }),
    })
  }
  return axios.get('/api/v1/assessments/autismdev/records/result-analysis/word', {
    params: { id },
    responseType: 'blob',
    headers: autismDevAuthHeaders(),
  })
}

export function downloadAutismDevIEPPlanWordApi(params: { id?: number | string, duration?: number | string, plan?: PEP3IEPPlanAIResult } = {}) {
  const headers = autismDevAuthHeaders()
  if (params.plan) {
    return axios.post('/api/v1/assessments/autismdev/records/iep-plan/word', {
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
  return axios.get('/api/v1/assessments/autismdev/records/iep-plan/word', {
    params,
    responseType: 'blob',
    headers,
  })
}

export function getAutismDevIEPPlanApi(id: number | string, durationMonths?: number | string) {
  return useGet<PEP3IEPPlanSavedVO>('/api/v1/assessments/autismdev/records/iep-plan/detail', {
    id,
    durationMonths: Number(durationMonths || 0),
  }, {
    loading: false,
    silentError: true,
  })
}

export function saveAutismDevIEPPlanApi(data: {
  id?: number | string
  durationMonths?: number | string
  status?: string
  resetExecutionPlans?: boolean
  plan: PEP3IEPPlanAIResult
}) {
  return usePost<PEP3IEPPlanSavedVO>('/api/v1/assessments/autismdev/records/iep-plan/save', {
    ...data,
    id: Number(data.id || 0),
    durationMonths: Number(data.durationMonths || 0),
  }, {
    loading: false,
    silentError: true,
    timeout: 60000,
  })
}

export function syncAutismDevIEPPlanPeriodApi(data: {
  id?: number | string
  durationMonths?: number | string
  sourceDurationMonths?: number | string
  startDate?: string
  startMonth?: string
}) {
  return usePost<PEP3IEPPlanPeriodSyncVO>('/api/v1/assessments/autismdev/records/iep-plan/period/sync', {
    ...data,
    id: Number(data.id || 0),
    durationMonths: Number(data.durationMonths || 0),
    sourceDurationMonths: Number(data.sourceDurationMonths || 0),
  }, {
    loading: false,
    silentError: true,
    timeout: 60000,
  })
}

export async function generateAutismDevIEPPlanAIStreamApi(
  data: { id?: number | string, durationMonths?: number | string },
  handlers: AutismDevIEPPlanAIStreamHandlers = {},
  options: PEP3IEPPlanAIStreamOptions = {},
) {
  const response = await fetch('/api/v1/assessments/autismdev/records/iep-plan/ai/stream', {
    method: 'POST',
    headers: autismDevStreamHeaders(),
    body: JSON.stringify({
      ...data,
      id: Number(data.id || 0),
      durationMonths: Number(data.durationMonths || 0),
    }),
    signal: options.signal,
  })
  return readAutismDevSSE<PEP3IEPPlanAIResult>(
    response,
    handlers,
    options,
    'AI生成失败',
    'AI生成未返回计划数据',
  )
}

export async function generateAutismDevExecutionPlanAIStreamApi(
  data: {
    id?: number | string
    durationMonths?: number | string
    planType: 'monthly' | 'weekly'
    targetMonthIndex?: number | string
    targetWeekIndex?: number | string
    restWeekdays?: number[]
    sourcePlan: PEP3IEPPlanAIResult
    monthlyPlan?: PEP3MonthlyPlanAIResult | null
  },
  handlers: AutismDevExecutionPlanAIStreamHandlers = {},
  options: PEP3IEPPlanAIStreamOptions = {},
) {
  const response = await fetch('/api/v1/assessments/autismdev/records/iep-plan/execution/ai/stream', {
    method: 'POST',
    headers: autismDevStreamHeaders(),
    body: JSON.stringify({
      ...data,
      id: Number(data.id || 0),
      durationMonths: Number(data.durationMonths || 0),
      targetMonthIndex: Number(data.targetMonthIndex || 0),
      targetWeekIndex: Number(data.targetWeekIndex || 0),
    }),
    signal: options.signal,
  })
  return readAutismDevSSE<PEP3MonthlyPlanAIResult | PEP3WeeklyPlanAIResult>(
    response,
    handlers,
    options,
    'AI生成失败',
    'AI生成未返回计划数据',
  )
}

export function downloadAutismDevExecutionPlanWordApi(data: {
  id?: number | string
  planType: 'monthly' | 'weekly'
  monthlyPlan?: PEP3MonthlyPlanAIResult | null
  weeklyPlan?: PEP3WeeklyPlanAIResult | null
}) {
  return axios.post('/api/v1/assessments/autismdev/records/iep-plan/execution/word', {
    ...data,
    id: Number(data.id || 0),
  }, {
    responseType: 'blob',
    headers: autismDevAuthHeaders({
      'Content-Type': 'application/json',
    }),
  })
}

export function getAutismDevExecutionPlansApi(id: number | string, durationMonths?: number | string) {
  return useGet<PEP3ExecutionPlanSavedVO>('/api/v1/assessments/autismdev/records/iep-plan/execution/detail', {
    id,
    durationMonths: Number(durationMonths || 0),
  }, {
    loading: false,
    silentError: true,
  })
}

export function saveAutismDevExecutionPlanApi(data: {
  id?: number | string
  durationMonths?: number | string
  planType: 'monthly' | 'weekly'
  targetMonthIndex?: number | string
  targetWeekIndex?: number | string
  monthlyPlan?: PEP3MonthlyPlanAIResult | null
  weeklyPlan?: PEP3WeeklyPlanAIResult | null
  preserveWeeklyPlans?: boolean
}) {
  return usePost<PEP3ExecutionPlanSavedVO>('/api/v1/assessments/autismdev/records/iep-plan/execution/save', {
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
