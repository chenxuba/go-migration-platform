import { useGet, usePost } from '~/utils/request'
import type {
  PageRequestModel,
  PageResult,
  PEP3AssessmentRecordQueryModel,
  PEP3AssessmentRecordSummary,
  PEP3DomainProgress,
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
