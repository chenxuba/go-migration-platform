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
