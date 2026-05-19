import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { useGet, usePost } from '~/utils/request'
import type {
  PageResult,
  PEP3AssessmentDraftProgress,
  PEP3AssessmentDraftQueryModel,
  PEP3AssessmentDraftSubmitResult,
  PEP3ExecutionPlanSavedVO,
  PEP3AssessmentRecordDetail,
  PEP3AssessmentRecordSummary,
  PEP3IEPPlanAIResult,
  PEP3IEPPlanAIStreamHandlers,
  PEP3IEPPlanAIStreamOptions,
  PEP3IEPPlanPeriodSyncVO,
  PEP3IEPPlanSavedVO,
  PEP3MonthlyPlanAIResult,
  PEP3RecordConfigUpdateRequest,
  PEP3RecordPageRequest,
  PEP3WeeklyPlanAIResult,
} from './pep3-assessment'

export type ShuangxiARecordPageRequest = PEP3RecordPageRequest
export type ShuangxiAAssessmentRecordSummary = PEP3AssessmentRecordSummary
export type ShuangxiAAssessmentRecordDetail = PEP3AssessmentRecordDetail
export type ShuangxiARecordConfigUpdateRequest = PEP3RecordConfigUpdateRequest
export type ShuangxiASelectedReportSection = 'developmentProfile' | 'resultAnalysis'

export interface ShuangxiAScoreOption {
  value: number
  label: string
  description?: string
}

export interface ShuangxiAItemSummary {
  itemNo: number
  itemCode?: string
  itemTitle?: string
  testItem?: string
  domainCode: string
  domainName: string
  skillCode: string
  skillName: string
}

export interface ShuangxiASkillSummary {
  skillCode: string
  skillName: string
  domainCode: string
  domainName: string
  sortNo?: number
  itemCount: number
  items: ShuangxiAItemSummary[]
}

export interface ShuangxiADomainSummary {
  domainCode: string
  domainName: string
  sortNo?: number
  itemCount: number
  maxRawScore?: number
  skills: ShuangxiASkillSummary[]
}

export interface ShuangxiATemplateSummary {
  title: string
  itemCount: number
  domainCount: number
  skillCount: number
  scoreMin?: number
  scoreMax?: number
  domains: ShuangxiADomainSummary[]
  scoreOptions: ShuangxiAScoreOption[]
}

export interface ShuangxiAAssessmentItem extends ShuangxiAItemSummary {
  scoreOptions: ShuangxiAScoreOption[]
}

export interface ShuangxiAItemScoreInput {
  itemNo: number
  score: number
  remark?: string
}

export interface ShuangxiAItemRemarkInput {
  itemNo: number
  remark: string
}

export interface ShuangxiADraftInput {
  studentId?: number
  studentName?: string
  studentGender?: string
  examinerName?: string
  remark?: string
  birthDate?: string
  assessmentDate?: string
  itemScores?: Record<number, number>
  itemScoreList?: ShuangxiAItemScoreInput[]
  itemRemarks?: Record<number, string>
  itemRemarkList?: ShuangxiAItemRemarkInput[]
}

export interface ShuangxiADraftSaveRequest extends ShuangxiADraftInput {
  id?: number
}

export interface ShuangxiADraftItemSaveRequest {
  draftId: number
  itemNo: number
  score: number
  remark?: string
  studentGender?: string
}

export interface ShuangxiADraftSummary {
  id: number
  instId?: number
  studentId?: number
  studentName?: string
  assessmentCode: string
  assessmentName: string
  scaleVersion?: string
  birthDate?: string
  assessmentDate?: string
  examinerId?: number
  examinerName?: string
  status: string
  submittedRecordId?: number
  answeredItemCount: number
  completionPercent: number
  progress: PEP3AssessmentDraftProgress
  remark?: string
  createdTime?: string
  updatedTime?: string
}

export interface ShuangxiADraftDetail extends ShuangxiADraftSummary {
  input?: ShuangxiADraftInput
}

export interface ShuangxiADraftPageRequest {
  pageRequestModel: {
    pageIndex: number
    pageSize: number
  }
  queryModel?: PEP3AssessmentDraftQueryModel
  latestOnly?: boolean
}

export interface ShuangxiARecordUpdateRequest extends ShuangxiADraftInput {
  id: number
}

export interface ShuangxiAResultAnalysisRow {
  domainCode?: string
  domain: string
  strengths: string
  weaknesses: string
  reason: string
  strategy: string
}

export interface ShuangxiAResultAnalysis {
  title: string
  courseName?: string
  model?: string
  generatedBy?: string
  generatedAt?: string
  rows: ShuangxiAResultAnalysisRow[]
}

export interface ShuangxiAResultAnalysisStreamHandlers {
  onStatus?: (message: string) => void
  onDelta?: (text: string) => void
  onDone?: (data: ShuangxiAResultAnalysis) => void
}

function normalizeShuangxiARecordPageRequest(data: ShuangxiARecordPageRequest): ShuangxiARecordPageRequest {
  const normalized = {
    ...data,
    queryModel: {
      ...data.queryModel,
      assessmentCode: 'SHUANGXI_A',
    },
  }
  const rawStudentId = normalized.queryModel?.studentId
  if (rawStudentId !== undefined && rawStudentId !== null && `${rawStudentId}`.trim() !== '') {
    const studentId = Number(rawStudentId)
    if (Number.isFinite(studentId) && studentId > 0)
      normalized.queryModel.studentId = studentId
  }
  return normalized
}

export function pageShuangxiAAssessmentRecordsApi(data: ShuangxiARecordPageRequest) {
  return usePost<PageResult<ShuangxiAAssessmentRecordSummary>>(
    '/api/v1/assessments/shuangxi-a/records/page',
    normalizeShuangxiARecordPageRequest(data),
    { silentError: true },
  )
}

export function getShuangxiAAssessmentFormTemplateSummaryApi() {
  return useGet<ShuangxiATemplateSummary>('/api/v1/assessments/shuangxi-a/form-template/summary')
}

export function getShuangxiAAssessmentFormTemplateItemApi(itemNo: number) {
  return useGet<ShuangxiAAssessmentItem>('/api/v1/assessments/shuangxi-a/form-template/item', { itemNo })
}

export function saveShuangxiAAssessmentDraftApi(data: ShuangxiADraftSaveRequest) {
  return usePost<ShuangxiADraftDetail>('/api/v1/assessments/shuangxi-a/drafts/save', data)
}

export function saveShuangxiAAssessmentDraftItemApi(data: ShuangxiADraftItemSaveRequest) {
  return usePost<ShuangxiADraftDetail>('/api/v1/assessments/shuangxi-a/drafts/item/save', data)
}

export function getShuangxiAAssessmentDraftDetailApi(id: number) {
  return useGet<ShuangxiADraftDetail>('/api/v1/assessments/shuangxi-a/drafts/detail', { id })
}

export function pageShuangxiAAssessmentDraftsApi(data: ShuangxiADraftPageRequest) {
  return usePost<PageResult<ShuangxiADraftSummary>>('/api/v1/assessments/shuangxi-a/drafts/page', data, { silentError: true })
}

export function submitShuangxiAAssessmentDraftApi(id: number) {
  return usePost<PEP3AssessmentDraftSubmitResult>('/api/v1/assessments/shuangxi-a/drafts/submit', { id })
}

export function getShuangxiAAssessmentRecordDetailApi(id: number) {
  return useGet<ShuangxiAAssessmentRecordDetail>('/api/v1/assessments/shuangxi-a/records/detail', { id })
}

export function updateShuangxiAAssessmentRecordApi(data: ShuangxiARecordUpdateRequest) {
  return usePost<ShuangxiAAssessmentRecordDetail>('/api/v1/assessments/shuangxi-a/records/update', data)
}

export function updateShuangxiAAssessmentRecordConfigApi(data: ShuangxiARecordConfigUpdateRequest) {
  return usePost<ShuangxiAAssessmentRecordDetail>('/api/v1/assessments/shuangxi-a/records/config/update', data)
}

export function deleteShuangxiAAssessmentRecordApi(id: number) {
  return usePost<boolean>('/api/v1/assessments/shuangxi-a/records/delete', { id })
}

export function downloadShuangxiADevelopmentProfilePdfApi(id: number) {
  return axios.get('/api/v1/assessments/shuangxi-a/records/development-profile/pdf', {
    params: { id },
    responseType: 'blob',
    headers: shuangxiAAuthHeaders(),
  })
}

export function downloadShuangxiASelectedReportPdfApi(
  id: number | string,
  sections: ShuangxiASelectedReportSection[] = ['developmentProfile'],
  analysis?: ShuangxiAResultAnalysis | null,
) {
  return axios.post('/api/v1/assessments/shuangxi-a/records/selected-report/pdf', {
    id: Number(id || 0),
    sections,
    ...(analysis ? { analysis } : {}),
  }, {
    responseType: 'blob',
    headers: shuangxiAAuthHeaders({
      'Content-Type': 'application/json',
    }),
  })
}

export function getShuangxiAResultAnalysisApi(id: number | string) {
  return useGet<ShuangxiAResultAnalysis>('/api/v1/assessments/shuangxi-a/records/result-analysis', { id }, {
    loading: false,
    silentError: true,
  })
}

export function saveShuangxiAResultAnalysisApi(id: number | string, analysis: ShuangxiAResultAnalysis) {
  return usePost<ShuangxiAResultAnalysis>('/api/v1/assessments/shuangxi-a/records/result-analysis', {
    id: Number(id || 0),
    analysis,
  }, {
    loading: false,
    silentError: true,
    timeout: 60000,
  })
}

function shuangxiAAuthHeaders(extra: Record<string, string> = {}) {
  const token = useAuthorization()
  return {
    [STORAGE_AUTHORIZE_KEY]: token.value || '',
    Authorization: token.value ? `Bearer ${token.value}` : '',
    'Accept-Language': 'zh-CN',
    ...extra,
  }
}

function shuangxiAStreamHeaders() {
  const headers: Record<string, string> = shuangxiAAuthHeaders({
    'Content-Type': 'application/json',
  })
  if (typeof window !== 'undefined')
    headers['X-Tenant-Domain'] = window.location.hostname.toLowerCase()
  return headers
}

async function readShuangxiSSE<T>(
  response: Response,
  handlers: {
    onStatus?: (message: string) => void
    onDelta?: (text: string) => void
    onDone?: (data: T) => void
  },
  options: PEP3IEPPlanAIStreamOptions,
  errorMessage: string,
  missingResultMessage: string,
) {
  if (!response.ok) {
    const text = await response.text()
    let message = text || errorMessage
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
      throw new Error(payload.message || errorMessage)
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
    throw new Error(missingResultMessage)
  return finalData
}

export function downloadShuangxiAIEPPlanWordApi(params: { id?: number | string, duration?: number | string, plan?: PEP3IEPPlanAIResult } = {}) {
  const headers = shuangxiAAuthHeaders()
  if (params.plan) {
    return axios.post('/api/v1/assessments/shuangxi-a/records/iep-plan/word', {
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
  return axios.get('/api/v1/assessments/shuangxi-a/records/iep-plan/word', {
    params,
    responseType: 'blob',
    headers,
  })
}

export function getShuangxiAIEPPlanApi(id: number | string, durationMonths?: number | string) {
  return useGet<PEP3IEPPlanSavedVO>('/api/v1/assessments/shuangxi-a/records/iep-plan/detail', {
    id,
    durationMonths: Number(durationMonths || 0),
  }, {
    loading: false,
    silentError: true,
  })
}

export function saveShuangxiAIEPPlanApi(data: {
  id?: number | string
  durationMonths?: number | string
  status?: string
  resetExecutionPlans?: boolean
  plan: PEP3IEPPlanAIResult
}) {
  return usePost<PEP3IEPPlanSavedVO>('/api/v1/assessments/shuangxi-a/records/iep-plan/save', {
    ...data,
    id: Number(data.id || 0),
    durationMonths: Number(data.durationMonths || 0),
  }, {
    loading: false,
    silentError: true,
    timeout: 60000,
  })
}

export function syncShuangxiAIEPPlanPeriodApi(data: {
  id?: number | string
  durationMonths?: number | string
  sourceDurationMonths?: number | string
  startDate?: string
  startMonth?: string
}) {
  return usePost<PEP3IEPPlanPeriodSyncVO>('/api/v1/assessments/shuangxi-a/records/iep-plan/period/sync', {
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

export async function generateShuangxiAIEPPlanAIStreamApi(
  data: { id?: number | string, durationMonths?: number | string },
  handlers: PEP3IEPPlanAIStreamHandlers = {},
  options: PEP3IEPPlanAIStreamOptions = {},
) {
  const response = await fetch('/api/v1/assessments/shuangxi-a/records/iep-plan/ai/stream', {
    method: 'POST',
    headers: shuangxiAStreamHeaders(),
    body: JSON.stringify({
      ...data,
      id: Number(data.id || 0),
      durationMonths: Number(data.durationMonths || 0),
    }),
    signal: options.signal,
  })
  return readShuangxiSSE<PEP3IEPPlanAIResult>(
    response,
    handlers,
    options,
    'AI生成失败',
    'AI生成未返回计划数据',
  )
}

export async function generateShuangxiAResultAnalysisStreamApi(
  id: number | string,
  handlers: ShuangxiAResultAnalysisStreamHandlers = {},
  options: PEP3IEPPlanAIStreamOptions = {},
) {
  const response = await fetch('/api/v1/assessments/shuangxi-a/records/result-analysis/ai/stream', {
    method: 'POST',
    headers: shuangxiAStreamHeaders(),
    body: JSON.stringify({ id: Number(id || 0) }),
    signal: options.signal,
  })
  return readShuangxiSSE<ShuangxiAResultAnalysis>(
    response,
    handlers,
    options,
    '评量结果分析生成失败',
    '评量结果分析生成未返回结果',
  )
}

export async function generateShuangxiAExecutionPlanAIStreamApi(
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
  handlers: {
    onStatus?: (message: string) => void
    onDelta?: (text: string) => void
    onDone?: (data: PEP3MonthlyPlanAIResult | PEP3WeeklyPlanAIResult) => void
  } = {},
  options: PEP3IEPPlanAIStreamOptions = {},
) {
  const response = await fetch('/api/v1/assessments/shuangxi-a/records/iep-plan/execution/ai/stream', {
    method: 'POST',
    headers: shuangxiAStreamHeaders(),
    body: JSON.stringify({
      ...data,
      id: Number(data.id || 0),
      durationMonths: Number(data.durationMonths || 0),
      targetMonthIndex: Number(data.targetMonthIndex || 0),
      targetWeekIndex: Number(data.targetWeekIndex || 0),
    }),
    signal: options.signal,
  })
  return readShuangxiSSE<PEP3MonthlyPlanAIResult | PEP3WeeklyPlanAIResult>(
    response,
    handlers,
    options,
    'AI生成失败',
    'AI生成未返回计划数据',
  )
}

export function downloadShuangxiAExecutionPlanWordApi(data: {
  id?: number | string
  planType: 'monthly' | 'weekly'
  monthlyPlan?: PEP3MonthlyPlanAIResult | null
  weeklyPlan?: PEP3WeeklyPlanAIResult | null
}) {
  return axios.post('/api/v1/assessments/shuangxi-a/records/iep-plan/execution/word', {
    ...data,
    id: Number(data.id || 0),
  }, {
    responseType: 'blob',
    headers: shuangxiAAuthHeaders({
      'Content-Type': 'application/json',
    }),
  })
}

export function getShuangxiAExecutionPlansApi(id: number | string, durationMonths?: number | string) {
  return useGet<PEP3ExecutionPlanSavedVO>('/api/v1/assessments/shuangxi-a/records/iep-plan/execution/detail', {
    id,
    durationMonths: Number(durationMonths || 0),
  }, {
    loading: false,
    silentError: true,
  })
}

export function saveShuangxiAExecutionPlanApi(data: {
  id?: number | string
  durationMonths?: number | string
  planType: 'monthly' | 'weekly'
  targetMonthIndex?: number | string
  targetWeekIndex?: number | string
  monthlyPlan?: PEP3MonthlyPlanAIResult | null
  weeklyPlan?: PEP3WeeklyPlanAIResult | null
  preserveWeeklyPlans?: boolean
}) {
  return usePost<PEP3ExecutionPlanSavedVO>('/api/v1/assessments/shuangxi-a/records/iep-plan/execution/save', {
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
