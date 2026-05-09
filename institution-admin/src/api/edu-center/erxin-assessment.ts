import axios from 'axios'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { useGet, usePost } from '~/utils/request'
import type { PageRequestModel, PageResult, PEP3AssessmentDraftProgress, PEP3AssessmentRecordDetail, PEP3AssessmentRecordQueryModel, PEP3AssessmentRecordSummary } from './pep3-assessment'

export type ERXinScaleCode = 'ERXIN2' | string

export interface ERXinScoreOption {
  value: boolean
  label: string
  description?: string
}

export interface ERXinAgeBand {
  ageMonth: number
  segment: string
  domainTotalScore: number
}

export interface ERXinAssessmentDomain {
  domainCode: string
  domainName: string
  sortNo: number
}

export interface ERXinAssessmentItemSummary {
  itemNo: number
  itemTitle: string
  testItem: string
  ageMonth: number
  ageSegment: string
  domainCode: string
  domainName: string
  parentReportAllowed: boolean
  attentionIfFailed: boolean
}

export interface ERXinAssessmentItem extends ERXinAssessmentItemSummary {
  domainMonthTotalScore: number
  itemWeight: number
  method: string
  passCriteria: string
  sourcePdf?: string
  sourcePages?: number[]
  ocrStatus?: string
}

export interface ERXinAssessmentAgeGroupSummary {
  groupCode: string
  title: string
  ageMonth: number
  segment: string
  domainTotalScore: number
  items: ERXinAssessmentItemSummary[]
}

export interface ERXinSubmitContract {
  scoreEndpoint: string
  createRecordEndpoint?: string
  dateFormat: string
  itemPassListKey: string
  requiredBaseFields: string[]
  allowedItemPassValues: boolean[]
}

export interface ERXinAssessmentFormTemplateSummary {
  templateCode: string
  templateVersion: string
  title: string
  scaleCode: ERXinScaleCode
  scaleVersion: string
  sourceStandard?: string
  sourcePdf?: string
  dataStatus?: string
  sources?: string[]
  itemCount: number
  ageBands: ERXinAgeBand[]
  scoreOptions: ERXinScoreOption[]
  domains: ERXinAssessmentDomain[]
  ageGroups: ERXinAssessmentAgeGroupSummary[]
  submitContract: ERXinSubmitContract
}

export interface ERXinItemPassInput {
  itemNo: number
  passed: boolean
  remark?: string
}

export interface ERXinItemRemarkInput {
  itemNo: number
  remark: string
}

export interface ERXinDraftSaveRequest {
  id?: number
  studentId?: number
  studentName?: string
  examinerName?: string
  remark?: string
  birthDate?: string
  assessmentDate?: string
  itemPasses?: Record<number, boolean>
  itemPassList?: ERXinItemPassInput[]
  itemRemarks?: Record<number, string>
  itemRemarkList?: ERXinItemRemarkInput[]
}

export interface ERXinDraftItemSaveRequest {
  draftId: number
  itemNo: number
  passed: boolean
  remark?: string
}

export interface ERXinDraftInput {
  studentId?: number
  studentName?: string
  examinerName?: string
  remark?: string
  birthDate?: string
  assessmentDate?: string
  itemPasses?: Record<number, boolean>
  itemPassList?: ERXinItemPassInput[]
  itemRemarks?: Record<number, string>
  itemRemarkList?: ERXinItemRemarkInput[]
}

export interface ERXinAssessmentDraftSummary {
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

export interface ERXinAssessmentDraftDetail extends ERXinAssessmentDraftSummary {
  input?: ERXinDraftInput
}

export interface ERXinAssessmentDraftQueryModel extends PEP3AssessmentRecordQueryModel {
  status?: 'draft' | 'ready_to_score' | 'complete' | 'submitted' | string
  latestOnly?: boolean
}

export interface ERXinDraftPageRequest {
  pageRequestModel: PageRequestModel
  queryModel: ERXinAssessmentDraftQueryModel
  latestOnly?: boolean
}

export interface ERXinAssessmentDraftSubmitResult {
  draftId: number
  recordId: number
  draftStatus: string
  record: PEP3AssessmentRecordDetail
}

export interface ERXinRecordPageRequest {
  pageRequestModel: PageRequestModel
  queryModel: PEP3AssessmentRecordQueryModel
}

export type ERXinAssessmentRecordSummary = PEP3AssessmentRecordSummary

export interface ERXinAssessmentRecordDetail extends ERXinAssessmentRecordSummary {
  input?: ERXinDraftInput
}

export interface ERXinRecordConfigUpdateRequest {
  id: number
  examinerName: string
  assessmentDate: string
}

export interface ERXinRecordUpdateRequest extends ERXinDraftSaveRequest {
  id: number
}

export interface ERXinReportInterpretation {
  title: string
  model?: string
  generatedBy: string
  generatedAt?: string
  summary: string
  domainAnalysis: string[]
  suggestions: string[]
  notes?: string[]
}

export interface ERXinReportInterpretationStreamHandlers {
  onStatus?: (message: string) => void
  onDelta?: (text: string) => void
  onDone?: (data: ERXinReportInterpretation) => void
}

export interface ERXinReportInterpretationStreamOptions {
  signal?: AbortSignal
}

function normalizeERXinPageRequest<T extends ERXinDraftPageRequest | ERXinRecordPageRequest>(data: T): T {
  const normalized = {
    ...data,
    queryModel: {
      ...data.queryModel,
      assessmentCode: 'ERXIN2',
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

export function getERXinAssessmentFormTemplateSummaryApi() {
  return useGet<ERXinAssessmentFormTemplateSummary>('/api/v1/assessments/erxin/form-template/summary')
}

export function getERXinAssessmentFormTemplateItemApi(itemNo: number) {
  return useGet<ERXinAssessmentItem>('/api/v1/assessments/erxin/form-template/item', { itemNo })
}

export function saveERXinAssessmentDraftApi(data: ERXinDraftSaveRequest) {
  return usePost<ERXinAssessmentDraftDetail>('/api/v1/assessments/erxin/drafts/save', data)
}

export function saveERXinAssessmentDraftItemApi(data: ERXinDraftItemSaveRequest) {
  return usePost<ERXinAssessmentDraftDetail>('/api/v1/assessments/erxin/drafts/item/save', data)
}

export function getERXinAssessmentDraftDetailApi(id: number) {
  return useGet<ERXinAssessmentDraftDetail>('/api/v1/assessments/erxin/drafts/detail', { id })
}

export function pageERXinAssessmentDraftsApi(data: ERXinDraftPageRequest) {
  return usePost<PageResult<ERXinAssessmentDraftSummary>>('/api/v1/assessments/erxin/drafts/page', normalizeERXinPageRequest(data), { silentError: true })
}

export function submitERXinAssessmentDraftApi(id: number) {
  return usePost<ERXinAssessmentDraftSubmitResult>('/api/v1/assessments/erxin/drafts/submit', { id })
}

export function pageERXinAssessmentRecordsApi(data: ERXinRecordPageRequest) {
  return usePost<PageResult<PEP3AssessmentRecordSummary>>('/api/v1/assessments/erxin/records/page', normalizeERXinPageRequest(data), { silentError: true })
}

export function getERXinAssessmentRecordDetailApi(id: number) {
  return useGet<ERXinAssessmentRecordDetail>('/api/v1/assessments/erxin/records/detail', { id })
}

export function updateERXinAssessmentRecordConfigApi(data: ERXinRecordConfigUpdateRequest) {
  return usePost<ERXinAssessmentRecordDetail>('/api/v1/assessments/erxin/records/config/update', data)
}

export function updateERXinAssessmentRecordApi(data: ERXinRecordUpdateRequest) {
  return usePost<ERXinAssessmentRecordDetail>('/api/v1/assessments/erxin/records/update', data)
}

export function deleteERXinAssessmentRecordApi(id: number) {
  return usePost<boolean>('/api/v1/assessments/erxin/records/delete', { id })
}

export function downloadERXinAssessmentRecordReportPdfApi(id: number) {
  const token = useAuthorization()
  return axios.get('/api/v1/assessments/erxin/records/report/pdf', {
    params: { id },
    responseType: 'blob',
    headers: {
      [STORAGE_AUTHORIZE_KEY]: token.value || '',
      Authorization: token.value ? `Bearer ${token.value}` : '',
      'Accept-Language': 'zh-CN',
    },
  })
}

export function getERXinAssessmentRecordReportInterpretationApi(id: number) {
  return useGet<ERXinReportInterpretation>('/api/v1/assessments/erxin/records/report/interpretation', { id }, { silentError: true })
}

export function generateERXinAssessmentRecordReportInterpretationApi(id: number) {
  return usePost<ERXinReportInterpretation>('/api/v1/assessments/erxin/records/report/interpretation/ai', { id }, {
    loading: false,
    silentError: true,
    timeout: 190000,
  })
}

function erxinStreamHeaders() {
  const token = useAuthorization()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    [STORAGE_AUTHORIZE_KEY]: token.value || '',
    Authorization: token.value ? `Bearer ${token.value}` : '',
    'Accept-Language': 'zh-CN',
    Accept: 'text/event-stream',
  }
  if (typeof window !== 'undefined')
    headers['X-Tenant-Domain'] = window.location.hostname.toLowerCase()
  return headers
}

export async function generateERXinAssessmentRecordReportInterpretationStreamApi(
  id: number,
  handlers: ERXinReportInterpretationStreamHandlers = {},
  options: ERXinReportInterpretationStreamOptions = {},
) {
  const response = await fetch('/api/v1/assessments/erxin/records/report/interpretation/ai/stream', {
    method: 'POST',
    headers: erxinStreamHeaders(),
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
  let finalInterpretation: ERXinReportInterpretation | null = null

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
