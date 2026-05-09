import {
  downloadPEP3ExecutionPlanWordApi,
  downloadPEP3IEPPlanWordApi,
  generatePEP3ExecutionPlanAIStreamApi,
  generatePEP3IEPPlanAIStreamApi,
  getPEP3AssessmentRecordReportInterpretationApi,
  getPEP3ExecutionPlansApi,
  getPEP3IEPPlanApi,
  savePEP3ExecutionPlanApi,
  savePEP3IEPPlanApi,
} from '@/api/edu-center/pep3-assessment'
import {
  downloadERXinExecutionPlanWordApi,
  downloadERXinIEPPlanWordApi,
  generateERXinExecutionPlanAIStreamApi,
  generateERXinIEPPlanAIStreamApi,
  getERXinAssessmentRecordReportInterpretationApi,
  getERXinExecutionPlansApi,
  getERXinIEPPlanApi,
  saveERXinExecutionPlanApi,
  saveERXinIEPPlanApi,
} from '@/api/edu-center/erxin-assessment'

const pep3MissingInterpretationConfirm = {
  title: '报告解读未生成',
  content: '您未生成报告解读，无法将报告解读资料用于本次AI生成IEP。点击确定后，将仅基于IEP教研库v3.0、PEP-3测评记录结果和儿童训练记录生成。',
  okText: '确定',
  cancelText: '取消',
}

const pep3Adapter = {
  key: 'PEP3',
  code: 'PEP3',
  aiLibraryLabel: 'IEP教研库v3.0',
  generationBasisText: 'PEP-3测评记录结果、报告解读和儿童训练记录',
  generationSourceText: 'IEP教研库v3.0、PEP-3测评记录结果、报告解读和儿童训练记录',
  generationFallbackBasisText: 'PEP-3测评记录结果和儿童训练记录',
  generationFallbackSourceText: 'IEP教研库v3.0、PEP-3测评记录结果和儿童训练记录',
  generationDescription: '正在读取PEP-3评估结果、报告解读和儿童训练记录，并生成可编辑的IEP表格。',
  emptyDescription: '点击“AI智能生成”后，系统会根据PEP-3测评结果、报告解读和近期训练记录实时生成表格。',
  getIepPlan: getPEP3IEPPlanApi,
  saveIepPlan: savePEP3IEPPlanApi,
  generateIepPlanStream: generatePEP3IEPPlanAIStreamApi,
  downloadIepPlanWord: downloadPEP3IEPPlanWordApi,
  getExecutionPlans: getPEP3ExecutionPlansApi,
  saveExecutionPlan: savePEP3ExecutionPlanApi,
  generateExecutionPlanStream: generatePEP3ExecutionPlanAIStreamApi,
  downloadExecutionPlanWord: downloadPEP3ExecutionPlanWordApi,
  missingInterpretationConfirm: pep3MissingInterpretationConfirm,
  async shouldConfirmBeforeGenerate(record?: any) {
    if (!record?.id)
      return null
    const response = await getPEP3AssessmentRecordReportInterpretationApi(record.id)
    return hasReportInterpretation(response) ? null : pep3MissingInterpretationConfirm
  },
}

const erxinMissingInterpretationConfirm = {
  title: '报告解读未生成',
  content: '您未生成报告解读，无法将报告解读资料用于本次AI生成IEP。点击确定后，将仅基于IEP教研库v3.0和儿心测评记录结果生成。',
  okText: '确定',
  cancelText: '取消',
}

const erxinAdapter = {
  key: 'ERXIN',
  code: 'ERXIN2',
  aiLibraryLabel: 'IEP教研库v3.0',
  generationBasisText: '儿心测评记录结果和报告解读',
  generationSourceText: 'IEP教研库v3.0、儿心测评记录结果和报告解读',
  generationFallbackBasisText: '儿心测评记录结果',
  generationFallbackSourceText: 'IEP教研库v3.0、儿心测评记录结果',
  generationDescription: '正在读取儿心评估结果和报告解读，并生成可编辑的IEP表格。',
  emptyDescription: '点击“AI智能生成”后，系统会根据儿心测评结果和报告解读实时生成表格。',
  missingInterpretationConfirm: erxinMissingInterpretationConfirm,
  async shouldConfirmBeforeGenerate(record: any) {
    if (!record?.id)
      return null
    const response = await getERXinAssessmentRecordReportInterpretationApi(record.id)
    return hasReportInterpretation(response) ? null : erxinMissingInterpretationConfirm
  },
  getIepPlan: getERXinIEPPlanApi,
  saveIepPlan: saveERXinIEPPlanApi,
  generateIepPlanStream: generateERXinIEPPlanAIStreamApi,
  downloadIepPlanWord: downloadERXinIEPPlanWordApi,
  getExecutionPlans: getERXinExecutionPlansApi,
  saveExecutionPlan: saveERXinExecutionPlanApi,
  generateExecutionPlanStream: generateERXinExecutionPlanAIStreamApi,
  downloadExecutionPlanWord: downloadERXinExecutionPlanWordApi,
}

const adapters = [pep3Adapter, erxinAdapter]

export type IEPPlanAssessmentAdapter = typeof pep3Adapter

function hasReportInterpretation(response: any) {
  const data = response?.data?.data || response?.data || response
  return !!(
    String(data?.summary || '').trim()
    || (Array.isArray(data?.domainAnalysis) && data.domainAnalysis.some((item: string) => String(item || '').trim()))
    || (Array.isArray(data?.suggestions) && data.suggestions.some((item: string) => String(item || '').trim()))
    || (Array.isArray(data?.notes) && data.notes.some((item: string) => String(item || '').trim()))
  )
}

export function resolveIEPPlanAssessmentAdapter(record: any): IEPPlanAssessmentAdapter {
  const source = String(record?._recordSource || record?.assessmentCode || '').trim().toUpperCase()
  return adapters.find(adapter => source === adapter.key || source === adapter.code || source.startsWith(adapter.key) || source.startsWith(adapter.code)) || pep3Adapter
}
