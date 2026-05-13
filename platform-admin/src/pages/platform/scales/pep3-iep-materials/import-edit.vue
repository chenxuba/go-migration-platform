<script setup lang="ts">
import { LeftOutlined } from '@ant-design/icons-vue'
import { Empty, Modal } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  batchSavePlatformPEP3IEPMaterialImportTaskRecordsApi,
  deletePlatformPEP3IEPMaterialImportTaskApi,
  getScaleQuestionBankApi,
  getPlatformPEP3IEPMaterialImportTaskDetailApi,
  getPlatformPEP3IEPMaterialImportTaskRecordListApi,
  startPlatformPEP3IEPMaterialImportTaskApi,
  type PEP3IEPMaterialImportColumn,
  type PEP3IEPMaterialImportRow,
  type ScaleQuestionBank,
  type ScaleQuestionBankItem,
} from '@/api/platform/scales'
import messageService from '@/utils/messageService'

const router = useRouter()
const route = useRoute()
const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE
const taskId = computed(() => String(route.params.id || ''))

const session = reactive({
  fileName: '',
  instName: '平台总控',
  columns: [] as PEP3IEPMaterialImportColumn[],
  rows: [] as PEP3IEPMaterialImportRow[],
  normalCount: 0,
  abnormalCount: 0,
})

const activeTab = ref<'normal' | 'abnormal'>('abnormal')
const taskLoading = ref(true)
const questionBankLoading = ref(false)
const deletingTask = ref(false)
const savingAll = ref(false)
const editModalOpen = ref(false)
const savingSingleCell = ref(false)
const editFormRef = ref()
const editModalState = reactive({
  rowId: '',
  cellKey: '',
  title: '',
  value: '',
})
const questionBank = ref<ScaleQuestionBank | null>(null)

const hasAbnormalRows = computed(() => session.abnormalCount > 0)
const optionMap = computed<Record<string, { label: string, value: string }[]>>(() => {
  const result: Record<string, { label: string, value: string }[]> = {}
  session.columns.forEach((column) => {
    if (column.options?.length) {
      result[column.title] = column.options.map(item => ({
        label: item,
        value: item,
      }))
    }
  })
  return result
})
const domainNameByCode = computed(() => {
  const result = new Map<string, string>()
  for (const item of questionBank.value?.domains || []) {
    const code = `${item.scaleCode || ''}`.trim()
    const name = `${item.scaleName || ''}`.trim()
    if (code && name)
      result.set(code, name)
  }
  return result
})
const allQuestionOptions = computed(() => {
  const items = questionBank.value?.items || []
  if (items.length > 0) {
    return items.map(item => ({
      label: questionTitle(item),
      value: questionTitle(item),
    }))
  }
  return optionMap.value['题目'] || []
})
const questionOptionsByDomainName = computed<Record<string, { label: string, value: string }[]>>(() => {
  const result: Record<string, { label: string, value: string }[]> = {}
  for (const item of questionBank.value?.items || []) {
    const domainName = questionDomainName(item)
    if (!domainName)
      continue
    if (!result[domainName])
      result[domainName] = []
    result[domainName].push({
      label: questionTitle(item),
      value: questionTitle(item),
    })
  }
  return result
})
const currentEditingContext = computed(() => findRowAndCell(editModalState.rowId, editModalState.cellKey))
const currentEditingColumn = computed(() => currentEditingContext.value.column)
const currentEditingOptions = computed(() => getColumnOptions(currentEditingContext.value.row, currentEditingContext.value.column))
const currentEditingUseSelect = computed(() => isSelectColumn(currentEditingContext.value.column))
const currentEditingDisabled = computed(() => isSelectDisabled(currentEditingContext.value.row, currentEditingContext.value.cell))
const currentEditingPlaceholder = computed(() => selectPlaceholder(currentEditingContext.value.row, currentEditingContext.value.cell))
const displayedRows = computed(() => {
  if (activeTab.value === 'normal')
    return session.rows.filter(row => !row.hasError)
  return session.rows.filter(row => row.hasError)
})
const tableMinWidth = computed(() => {
  return session.columns.reduce((total, column) => total + getColumnWidth(column.title), 0) + 70 + 90
})

function unwrap<T>(res: any): T {
  return (res?.result ?? res?.data ?? res) as T
}

function getColumnWidth(title: string) {
  switch (`${title || ''}`.trim()) {
    case '领域':
      return 150
    case '题目':
      return 260
    case '选项':
    case '状态':
      return 110
    case '课程形式':
      return 120
    case '训练项目':
      return 180
    case '长期目标':
    case '短期目标':
    case '训练内容':
      return 260
    default:
      return 160
  }
}

function questionTitle(item: ScaleQuestionBankItem) {
  return item.itemTitle || item.testItem || `第${item.itemNo || ''}题`
}

function questionDomainName(item: ScaleQuestionBankItem) {
  const domainCode = `${item.domainCode || ''}`.trim()
  return `${item.domainName || domainNameByCode.value.get(domainCode) || domainCode}`.trim()
}

function getRowCell(row: PEP3IEPMaterialImportRow | undefined, title: string) {
  return row?.cells?.find(cell => cell.title === title)
}

function getRowCellValue(row: PEP3IEPMaterialImportRow | undefined, title: string) {
  return `${getRowCell(row, title)?.value || ''}`.trim()
}

function isSelectColumn(column: PEP3IEPMaterialImportColumn | undefined) {
  if (!column)
    return false
  return column.fieldType === 4 || ['领域', '题目', '选项', '课程形式', '状态'].includes(column.title)
}

function getQuestionOptionsForRow(row: PEP3IEPMaterialImportRow | undefined) {
  const domainName = getRowCellValue(row, '领域')
  if (!domainName)
    return []
  return questionOptionsByDomainName.value[domainName] || []
}

function getColumnOptions(row: PEP3IEPMaterialImportRow | undefined, column: PEP3IEPMaterialImportColumn | undefined) {
  if (!column)
    return []
  if (column.title === '题目')
    return getQuestionOptionsForRow(row)
  return optionMap.value[column.title] || []
}

function getCellOptions(row: PEP3IEPMaterialImportRow, cell: any) {
  return getColumnOptions(row, getColumnByCell(cell))
}

function isSelectCell(row: PEP3IEPMaterialImportRow, cell: any) {
  return isSelectColumn(getColumnByCell(cell))
}

function isSelectDisabled(row: PEP3IEPMaterialImportRow | undefined, cell: any) {
  return cell?.title === '题目' && !getRowCellValue(row, '领域')
}

function selectPlaceholder(row: PEP3IEPMaterialImportRow | undefined, cell: any) {
  if (cell?.title === '题目' && !getRowCellValue(row, '领域'))
    return '请先选择领域'
  return '请选择'
}

function isQuestionMatchedDomain(row: PEP3IEPMaterialImportRow | undefined, questionName: string) {
  const text = `${questionName || ''}`.trim()
  if (!text)
    return true
  const domainName = getRowCellValue(row, '领域')
  if (!domainName)
    return false
  return getQuestionOptionsForRow(row).some(item => item.value === text || item.label === text)
}

function recomputeSummary() {
  session.normalCount = session.rows.filter(row => !row.hasError).length
  session.abnormalCount = session.rows.filter(row => row.hasError).length
  if (session.abnormalCount === 0)
    activeTab.value = 'normal'
}

function validateCell(row: PEP3IEPMaterialImportRow | undefined, column: PEP3IEPMaterialImportColumn | undefined, value: string) {
  if (!column)
    return ''
  const text = `${value || ''}`.trim()
  if (column.required && !text)
    return '请填写'
  if (!text)
    return ''
  if (column.title === '题目' && !isQuestionMatchedDomain(row, text))
    return '题目不属于所选领域'
  const options = getColumnOptions(row, column)
  if (options.length > 0 && !options.some(item => item.value === text || item.label === text))
    return '请选择预设值'
  return ''
}

function syncQuestionAfterDomainChange(row: PEP3IEPMaterialImportRow) {
  const questionCell = getRowCell(row, '题目')
  const questionColumn = session.columns.find(item => item.title === '题目')
  if (!questionCell || !questionColumn)
    return
  if (questionCell.value && !isQuestionMatchedDomain(row, questionCell.value)) {
    questionCell.value = ''
    questionCell.selectedId = undefined
  }
  questionCell.error = validateCell(row, questionColumn, questionCell.value)
}

function applyCellDraft(row: PEP3IEPMaterialImportRow, cell: any, column: PEP3IEPMaterialImportColumn | undefined, value: any, validateNow = false) {
  const text = `${value ?? ''}`.trim()
  const options = getColumnOptions(row, column)
  const matched = options.find(item => item.value === text || item.label === text)
  cell.value = matched ? matched.label : text
  cell.selectedId = matched ? matched.value : undefined
  cell.error = validateCell(row, column, cell.value)
  if (column?.title === '领域')
    syncQuestionAfterDomainChange(row)
  if (validateNow) {
    row.hasError = row.cells.some(item => item.error)
    recomputeSummary()
  }
}

function handleCellChange(row: PEP3IEPMaterialImportRow, cell: any, column: PEP3IEPMaterialImportColumn | undefined, value: any) {
  applyCellDraft(row, cell, column, value, false)
}

function getColumnByCell(cell: any) {
  return session.columns.find(column => column.key === cell.key)
}

function getDisplayCellText(cell: any) {
  const text = `${cell?.value || ''}`.trim()
  return text || '-'
}

function findRowAndCell(rowId: string, cellKey: string) {
  const row = session.rows.find(item => item.id === rowId)
  const cell = row?.cells.find(item => item.key === cellKey)
  const column = session.columns.find(item => item.key === cellKey)
  return { row, cell, column }
}

function openEditModal(row: PEP3IEPMaterialImportRow, cell: any) {
  editModalState.rowId = row.id
  editModalState.cellKey = cell.key
  editModalState.title = cell.title
  editModalState.value = cell.value || ''
  editModalOpen.value = true
}

async function handleConfirmEditModal() {
  const { row, cell, column } = findRowAndCell(editModalState.rowId, editModalState.cellKey)
  if (!row || !cell || !column)
    return
  const error = validateCell(row, column, editModalState.value)
  if (error) {
    messageService.warning(error)
    return
  }

  applyCellDraft(row, cell, column, editModalState.value, false)
  savingSingleCell.value = true
  try {
    const rows = unwrap<PEP3IEPMaterialImportRow[]>(await batchSavePlatformPEP3IEPMaterialImportTaskRecordsApi({
      taskId: taskId.value,
      records: [row],
    }))
    const rowMap = new Map(rows.map(item => [item.id, item]))
    session.rows = session.rows.map(item => rowMap.get(item.id) || item)
    recomputeSummary()
    editModalOpen.value = false
    messageService.success('保存成功')
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '保存失败')
  } finally {
    savingSingleCell.value = false
  }
}

function handleDeleteRow(rowNo: number) {
  session.rows = session.rows.filter(row => row.rowNo !== rowNo)
  recomputeSummary()
}

async function handleSave() {
  savingAll.value = true
  try {
    const rows = unwrap<PEP3IEPMaterialImportRow[]>(await batchSavePlatformPEP3IEPMaterialImportTaskRecordsApi({
      taskId: taskId.value,
      records: session.rows,
    }))
    const rowMap = new Map(rows.map(item => [item.id, item]))
    session.rows = session.rows.map(row => rowMap.get(row.id) || row)
    recomputeSummary()
    messageService.success('已保存修改')
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '保存修改失败')
  } finally {
    savingAll.value = false
  }
}

async function handleStartImport() {
  recomputeSummary()
  if (session.abnormalCount > 0) {
    messageService.warning('请先处理异常数据')
    activeTab.value = 'abnormal'
    return
  }
  try {
    await startPlatformPEP3IEPMaterialImportTaskApi({ taskId: taskId.value })
    messageService.success('开始导入，请稍后')
    router.push('/platform/scales/pep3-iep-materials/import/record')
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '开始导入失败')
  }
}

function handleBack() {
  router.replace('/platform/scales/pep3-iep-materials/import')
}

function handleCancel() {
  Modal.confirm({
    title: '确认取消导入并返回？',
    centered: true,
    okText: '确认取消',
    okType: 'danger',
    cancelText: '继续处理',
    content: '取消后会删除当前这条导入任务记录。',
    async onOk() {
      deletingTask.value = true
      try {
        await deletePlatformPEP3IEPMaterialImportTaskApi({ taskId: taskId.value })
        messageService.success('已取消导入并删除记录')
        router.replace('/platform/scales/pep3-iep-materials/import')
      } catch (error: any) {
        messageService.error(error?.response?.data?.message || error?.message || '取消导入失败')
        return Promise.reject(error)
      } finally {
        deletingTask.value = false
      }
    },
  })
}

async function loadTaskData() {
  const [detailRes, abnormalRes, normalRes] = await Promise.all([
    getPlatformPEP3IEPMaterialImportTaskDetailApi({ taskId: taskId.value }),
    getPlatformPEP3IEPMaterialImportTaskRecordListApi({
      queryModel: { taskId: taskId.value, type: 0 },
      pageRequestModel: { needTotal: true, pageSize: 1000, pageIndex: 1, skipCount: 0 },
    }),
    getPlatformPEP3IEPMaterialImportTaskRecordListApi({
      queryModel: { taskId: taskId.value, type: 1 },
      pageRequestModel: { needTotal: true, pageSize: 1000, pageIndex: 1, skipCount: 0 },
    }),
  ])
  const detail = unwrap<any>(detailRes)
  const abnormal = unwrap<any>(abnormalRes)
  const normal = unwrap<any>(normalRes)
  session.fileName = detail?.fileName || ''
  session.instName = detail?.instName || '平台总控'
  session.columns = abnormal?.columns?.length ? abnormal.columns : (normal?.columns || [])
  session.rows = [...(abnormal?.list || []), ...(normal?.list || [])]
  recomputeSummary()
  if (session.abnormalCount > 0)
    activeTab.value = 'abnormal'
}

async function loadQuestionBank() {
  questionBankLoading.value = true
  try {
    questionBank.value = unwrap<ScaleQuestionBank>(await getScaleQuestionBankApi({ scaleCode: 'PEP3' }))
  } catch (error: any) {
    questionBank.value = null
    messageService.error(error?.response?.data?.message || error?.message || '加载PEP3题库失败')
  } finally {
    questionBankLoading.value = false
  }
}

onMounted(() => {
  taskLoading.value = true
  Promise.all([loadTaskData(), loadQuestionBank()]).catch((error: any) => {
    messageService.error(error?.response?.data?.message || error?.message || '导入任务加载失败')
    router.replace('/platform/scales/pep3-iep-materials/import')
  }).finally(() => {
    taskLoading.value = false
  })
})
</script>

<template>
  <div class="import-edit-layout">
    <div class="work-top">
      <div class="work-top-left">
        <div class="import-header-logo" title="导入中心" aria-hidden="true" />
        <span class="back-link" @click="handleBack">
          <LeftOutlined /> 返回
        </span>
      </div>
      <div class="work-top-right">
        当前端：{{ session.instName || '平台总控' }}
      </div>
    </div>

    <div class="work-main">
      <div class="work-main-card">
        <div class="title-row">
          <div class="file-title">
            {{ session.fileName || '正在解析导入文件...' }}
          </div>
          <div class="actions">
            <a-button :loading="deletingTask" @click="handleCancel">
              取消导入并返回
            </a-button>
            <a-button type="primary" class="ml-12px" @click="handleStartImport">
              开始导入
            </a-button>
          </div>
        </div>

        <div v-if="taskLoading" class="task-loading-panel">
          <a-spin size="large" />
          <div class="task-loading-title">
            正在解析导入文件
          </div>
          <div class="task-loading-desc">
            数据量较大时可能需要几秒，请稍候。
          </div>
        </div>

        <template v-else>
          <a-alert
            v-if="hasAbnormalRows"
            class="mt-20px"
            type="warning"
            show-icon
            message="文件存在异常数据"
            description="请修改或删除异常数据，当异常数据全部处理完成后，可点击「开始导入」。"
          />

          <div class="tab-row">
            <div class="tabs">
              <span :class="['tab', { active: activeTab === 'normal' }]" @click="activeTab = 'normal'">正常({{ session.normalCount || 0 }})</span>
              <span :class="['tab', { active: activeTab === 'abnormal' }]" @click="activeTab = 'abnormal'">异常({{ session.abnormalCount || 0 }})</span>
            </div>
            <a-button v-if="hasAbnormalRows" type="primary" ghost :loading="savingAll" @click="handleSave">
              保存修改
            </a-button>
          </div>

          <div class="table-wrap">
            <table class="edit-table" :style="{ minWidth: `${tableMinWidth}px` }">
              <colgroup>
                <col style="width: 70px">
                <col v-for="column in session.columns" :key="column.key" :style="{ width: `${getColumnWidth(column.title)}px` }">
                <col style="width: 90px">
              </colgroup>
              <thead>
                <tr>
                  <th class="index-column">序号</th>
                  <th v-for="column in session.columns" :key="column.key">
                    <span v-if="column.required" class="required">*</span>{{ column.title }}
                  </th>
                  <th class="action-column">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in displayedRows" :key="row.id">
                  <td class="index-column">{{ row.rowNo }}</td>
                  <td v-for="cell in row.cells" :key="cell.key">
                    <template v-if="activeTab === 'abnormal' && isSelectCell(row, cell)">
                      <a-select
                        :value="cell.value || undefined"
                        allow-clear
                        show-search
                        option-filter-prop="label"
                        :disabled="isSelectDisabled(row, cell)"
                        :loading="questionBankLoading && cell.title === '题目'"
                        :options="getCellOptions(row, cell)"
                        :placeholder="selectPlaceholder(row, cell)"
                        style="width: 100%"
                        @change="value => handleCellChange(row, cell, getColumnByCell(cell), value)"
                      />
                    </template>
                    <template v-else-if="activeTab === 'abnormal'">
                      <a-input
                        :value="cell.value"
                        placeholder="请输入"
                        @input="event => handleCellChange(row, cell, getColumnByCell(cell), (event.target as HTMLInputElement).value)"
                      />
                    </template>
                    <template v-else>
                      <div class="readonly-cell" @click="openEditModal(row, cell)">
                        <span class="readonly-cell__text">{{ getDisplayCellText(cell) }}</span>
                        <span class="readonly-cell__edit">编辑</span>
                      </div>
                    </template>
                    <div v-if="cell.error" class="error-text">
                      {{ cell.error }}
                    </div>
                  </td>
                  <td class="action-column">
                    <a-button type="link" danger @click="handleDeleteRow(row.rowNo)">
                      删除
                    </a-button>
                  </td>
                </tr>
                <tr v-if="displayedRows.length === 0">
                  <td :colspan="session.columns.length + 2" class="empty-table-cell">
                    <a-empty :image="simpleImage" />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </div>
    </div>

    <a-modal
      v-model:open="editModalOpen"
      centered
      :confirm-loading="savingSingleCell"
      :title="`编辑${editModalState.title}`"
      ok-text="保存"
      cancel-text="取消"
      @ok="handleConfirmEditModal"
    >
      <a-form ref="editFormRef" layout="vertical" :model="editModalState">
        <a-form-item :label="editModalState.title" name="value">
          <a-select
            v-if="currentEditingUseSelect"
            v-model:value="editModalState.value"
            show-search
            allow-clear
            option-filter-prop="label"
            :disabled="currentEditingDisabled"
            :loading="questionBankLoading && editModalState.title === '题目'"
            :options="currentEditingOptions"
            :placeholder="currentEditingPlaceholder"
          />
          <a-textarea v-else v-model:value="editModalState.value" :rows="4" placeholder="请输入" />
        </a-form-item>
      </a-form>
      <div v-if="currentEditingColumn?.required" class="modal-tip">
        该字段为必填项。
      </div>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.import-edit-layout {
  height: 100vh;
  min-height: 100vh;
  overflow: hidden;
  background: #f7f7fd;
}

.work-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 52px;
  background: #fff;
}

.work-top-left {
  display: flex;
  align-items: center;
}

.work-top-right {
  padding-right: 24px;
  color: #000;
  font-size: 15px;
  font-weight: 500;
}

.import-header-logo {
  position: relative;
  width: 52px;
  height: 52px;
  flex-shrink: 0;
  overflow: hidden;
  background: linear-gradient(145deg, #2b8cff 0%, #0066ff 45%, #0050d8 100%);
}

.import-header-logo::before {
  position: absolute;
  top: 14px;
  left: 11px;
  width: 30px;
  height: 24px;
  background: #fff;
  box-shadow: inset 0 -8px 0 rgba(0, 102, 255, 0.12), inset 0 -16px 0 rgba(0, 102, 255, 0.08);
  content: '';
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: 14px;
  color: #06f;
  font-size: 18px;
  font-weight: 500;
  cursor: pointer;
}

.work-main {
  box-sizing: border-box;
  display: flex;
  justify-content: center;
  height: calc(100vh - 52px);
  min-width: 1040px;
  padding: 24px 40px;
}

.work-main-card {
  box-sizing: border-box;
  width: min(1260px, 100%);
  height: 100%;
  min-height: 0;
  overflow: hidden;
  padding: 28px 48px 30px;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 12px 32px rgba(15, 35, 80, 0.08);
}

.title-row,
.tab-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.file-title {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  color: #111827;
  font-size: 22px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.actions {
  flex-shrink: 0;
}

.task-loading-panel {
  display: flex;
  height: 480px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #667085;
}

.task-loading-title {
  margin-top: 18px;
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
}

.task-loading-desc {
  margin-top: 8px;
}

.tab-row {
  margin-top: 20px;
}

.tabs {
  display: flex;
  gap: 28px;
}

.tab {
  position: relative;
  padding: 8px 0;
  color: #667085;
  font-size: 16px;
  cursor: pointer;
}

.tab.active {
  color: #1677ff;
  font-weight: 600;
}

.tab.active::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 2px;
  border-radius: 2px;
  background: #1677ff;
  content: '';
}

.table-wrap {
  margin-top: 16px;
  max-height: calc(100vh - 262px);
  overflow: auto;
  border: 1px solid #eef0f4;
  border-radius: 8px;
}

.edit-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  table-layout: fixed;
}

.edit-table th {
  height: 44px;
  padding: 0 12px;
  border-bottom: 1px solid #e5e7eb;
  background: #f8fafc;
  color: #344054;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
  text-align: left;
}

.edit-table td {
  min-height: 52px;
  padding: 10px 12px;
  border-bottom: 1px solid #f0f2f5;
  color: #1f2937;
  font-size: 14px;
  font-weight: 400;
  line-height: 1.5;
  vertical-align: top;
}

.index-column,
.action-column {
  text-align: center !important;
}

.required {
  margin-right: 2px;
  color: #ff4d4f;
}

.error-text {
  margin-top: 5px;
  color: #ff4d4f;
  font-size: 12px;
}

.readonly-cell {
  display: flex;
  min-height: 32px;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  cursor: pointer;
}

.readonly-cell__text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.readonly-cell__edit {
  flex-shrink: 0;
  color: #1677ff;
  font-size: 12px;
  opacity: 0;
}

.readonly-cell:hover .readonly-cell__edit {
  opacity: 1;
}

.empty-table-cell {
  height: 240px;
}

.modal-tip {
  margin-top: -8px;
  color: #8a94a6;
  font-size: 12px;
}
</style>
