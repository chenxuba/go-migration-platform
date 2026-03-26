<script setup>
import { ExclamationCircleOutlined, LeftOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import { createVNode } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dayjs from 'dayjs'
import messageService from '~@/utils/messageService'
import { ParentRelationshipLabel, SexLabel } from '@/enums'
import { getStuCustomFieldApi, getStuDefaultFieldApi } from '@/api/edu-center/student-list'
import { getStaffSummariesApi } from '@/api/finance-center/approval-manage'
import { batchSaveIntentionStudentImportTaskRecordsApi, deleteIntentionStudentImportTaskApi, getChannelTreeApi, getIntentionStudentImportTaskDetailApi, getIntentionStudentImportTaskRecordListApi, startIntentionStudentImportTaskApi } from '~@/api/enroll-center/intention-student'

const router = useRouter()
const route = useRoute()
const importId = computed(() => String(route.params.id || ''))
const session = reactive({
  importId: importId.value,
  fileName: '',
  instName: '总机构',
  columns: [],
  rows: [],
  normalCount: 0,
  abnormalCount: 0,
})

const activeTab = ref('abnormal')
const optionMap = ref({})
const hasAbnormalRows = computed(() => session.abnormalCount > 0)
const taskLoading = ref(true)
const deletingTask = ref(false)
const editModalOpen = ref(false)
const savingSingleCell = ref(false)
const editFormRef = ref()
const editModalState = reactive({
  rowId: '',
  cellKey: '',
  title: '',
  fieldType: 1,
  value: '',
  selectedId: undefined,
})

const currentEditingOptions = computed(() => optionMap.value[editModalState.title] || [])
const currentEditingColumn = computed(() => session.columns.find(col => col.key === editModalState.cellKey) || null)

const phoneRelationshipOptions = [
  { label: ParentRelationshipLabel[1], value: '1' },
  { label: ParentRelationshipLabel[2], value: '2' },
  { label: ParentRelationshipLabel[3], value: '3' },
  { label: ParentRelationshipLabel[4], value: '4' },
  { label: ParentRelationshipLabel[5], value: '5' },
  { label: ParentRelationshipLabel[6], value: '6' },
  { label: ParentRelationshipLabel[7], value: '7' },
]

const sexOptions = [
  { label: SexLabel[1], value: '1' },
  { label: SexLabel[0], value: '0' },
  { label: SexLabel[2], value: '2' },
]

function getColumnWidth(title) {
  switch (`${title || ''}`.trim()) {
    case '学员姓名':
      return 140
    case '手机号':
      return 170
    case '手机号归属人':
      return 170
    case '渠道':
      return 170
    case '性别':
      return 110
    case '生日':
      return 150
    case '微信号':
      return 140
    case '年级':
      return 120
    case '就读学校':
      return 160
    case '家庭住址':
      return 180
    case '兴趣爱好':
      return 160
    case '残疾证号':
      return 150
    case '身份证号':
      return 170
    case '籍贯':
      return 140
    case '销售员':
      return 140
    default:
      return 150
  }
}

const tableMinWidth = computed(() => {
  const dynamicWidth = session.columns.reduce((total, column) => total + getColumnWidth(column.title), 0)
  return dynamicWidth + 70 + 90
})

const displayedRows = computed(() => {
  if (activeTab.value === 'normal')
    return session.rows.filter(row => !row.hasError)
  return session.rows.filter(row => row.hasError)
})

function recomputeSummary() {
  session.normalCount = session.rows.filter(row => !row.hasError).length
  session.abnormalCount = session.rows.filter(row => row.hasError).length
  if (session.abnormalCount === 0)
    activeTab.value = 'normal'
  else if (activeTab.value !== 'normal' && activeTab.value !== 'abnormal')
    activeTab.value = 'abnormal'
}

function setCellOptionSelection(cell, options = []) {
  const matched = options.find(item => `${item.label}` === `${cell.value || ''}` || `${item.value}` === `${cell.value || ''}`)
  cell.selectedId = matched ? matched.value : undefined
}

function refreshRowOptionSelections() {
  session.rows.forEach((row) => {
    row.cells.forEach((cell) => {
      setCellOptionSelection(cell, optionMap.value[cell.title] || [])
    })
  })
}

function validateCell(column, value) {
  const text = `${value || ''}`.trim()
  if (column.required && !text)
    return '请填写'
  if (!text)
    return ''
  const optionList = optionMap.value[column.title] || []
  if (optionList.length > 0 && !optionList.some(item => `${item.label}` === text || `${item.value}` === text))
    return '请选择预设值'
  if (column.title === '手机号' && !/^1\d{10}$/.test(text))
    return '手机号格式错误'
  if (column.fieldType === 2 && Number.isNaN(Number(text)))
    return '请输入数字'
  if (column.fieldType === 3) {
    const valid = [
      /^\d{4}-\d{1,2}-\d{1,2}$/,
      /^\d{4}\/\d{1,2}\/\d{1,2}$/,
      /^\d{4}\.\d{1,2}\.\d{1,2}$/,
      /^\d{4}年\d{1,2}月\d{1,2}日$/,
      /^\d{8}$/,
      /^\d{2}-\d{2}-\d{2}$/,
    ].some(reg => reg.test(text))
    if (!valid)
      return '日期格式错误'
  }
  return ''
}

function applyCellDraft(row, cell, column, value, validateNow = false) {
  const options = optionMap.value[column.title] || []
  const matched = options.find(item => `${item.value}` === `${value}` || `${item.label}` === `${value}`)
  cell.value = matched ? matched.label : value
  cell.selectedId = matched ? matched.value : undefined
  cell.error = validateCell(column, value)
  if (validateNow) {
    row.hasError = row.cells.some(item => item.error)
    recomputeSummary()
  }
}

function handleCellChange(row, cell, column, value) {
  applyCellDraft(row, cell, column, value, false)
}

function findRowAndCell(rowId, cellKey) {
  const row = session.rows.find(item => item.id === rowId)
  if (!row)
    return {}
  const cell = row.cells.find(item => item.key === cellKey)
  const column = session.columns.find(item => item.key === cellKey)
  return { row, cell, column }
}

function openEditModal(row, cell) {
  const column = session.columns.find(item => item.key === cell.key)
  if (!column)
    return
  editModalState.rowId = row.id
  editModalState.cellKey = cell.key
  editModalState.title = cell.title
  editModalState.fieldType = column.fieldType
  editModalState.value = cell.value || ''
  editModalState.selectedId = cell.selectedId
  editModalOpen.value = true
}

function getDisplayCellText(cell) {
  const text = `${cell?.value || ''}`.trim()
  return text || '-'
}

const editFieldRules = computed(() => {
  const column = currentEditingColumn.value
  if (!column)
    return []
  const rules = []
  if (column.required) {
    rules.push({
      required: true,
      validator: async () => {
        const options = currentEditingOptions.value
        const nextValue = options.length > 0
          ? `${editModalState.selectedId || ''}`.trim()
          : `${editModalState.value || ''}`.trim()
        if (!nextValue)
          return Promise.reject(new Error('请填写'))
        return Promise.resolve()
      },
      trigger: ['change', 'blur'],
    })
  }
  if (column.title === '手机号') {
    rules.push({
      validator: async () => {
        const nextValue = `${editModalState.value || ''}`.trim()
        if (!nextValue)
          return Promise.resolve()
        if (!/^1\d{10}$/.test(nextValue))
          return Promise.reject(new Error('手机号格式错误'))
        return Promise.resolve()
      },
      trigger: ['change', 'blur'],
    })
  }
  return rules
})

async function handleConfirmEditModal() {
  const { row, cell, column } = findRowAndCell(editModalState.rowId, editModalState.cellKey)
  if (!row || !cell || !column)
    return

  try {
    await editFormRef.value?.validate()
  }
  catch {
    return
  }

  const nextValue = currentEditingOptions.value.length > 0 ? editModalState.selectedId : editModalState.value
  applyCellDraft(row, cell, column, nextValue, false)

  savingSingleCell.value = true
  try {
    const { result, data } = await batchSaveIntentionStudentImportTaskRecordsApi({
      taskId: importId.value,
      records: [row],
    })
    const rows = result || data || []
    const rowMap = new Map(rows.map(item => [item.id, item]))
    session.rows = session.rows.map(item => rowMap.get(item.id) || item)
    recomputeSummary()
    refreshRowOptionSelections()
    editModalOpen.value = false
    messageService.success('保存成功')
  }
  catch (error) {
    console.error('save single import cell failed', error)
    messageService.error('保存失败')
  }
  finally {
    savingSingleCell.value = false
  }
}

function handleDeleteRow(rowNo) {
  session.rows = session.rows.filter(row => row.rowNo !== rowNo)
  recomputeSummary()
}

function handleBack() {
  router.replace('/import-center/import-intention-student-starter')
}

function handleSave() {
  batchSaveIntentionStudentImportTaskRecordsApi({
    taskId: importId.value,
    records: session.rows,
  }).then(({ result, data }) => {
    const rows = result || data || []
    const rowMap = new Map(rows.map(item => [item.id, item]))
    session.rows = session.rows.map(row => rowMap.get(row.id) || row)
    recomputeSummary()
    refreshRowOptionSelections()
    messageService.success('已保存修改')
  }).catch((error) => {
    console.error('save import task records failed', error)
    messageService.error('保存修改失败')
  })
}

function handleStartImport() {
  recomputeSummary()
  if (session.abnormalCount > 0) {
    messageService.warning('请先处理异常数据')
    activeTab.value = 'abnormal'
    return
  }
  startIntentionStudentImportTaskApi({ taskId: importId.value }).then(() => {
    messageService.success('开始导入，请稍后')
    router.push('/import-center/import-intention-student/record')
  }).catch((error) => {
    console.error('start intention student import failed', error)
    messageService.error(error?.message || '开始导入失败')
  })
}

function handleCancel() {
  Modal.confirm({
    title: '确认取消导入并返回？',
    icon: createVNode(ExclamationCircleOutlined),
    centered: true,
    okText: '确认取消',
    okType: 'danger',
    cancelText: '继续处理',
    content: '取消后会删除当前这条导入任务记录，该操作不可恢复。',
    async onOk() {
      if (!importId.value) {
        router.replace('/import-center/import-intention-student-starter')
        return
      }
      deletingTask.value = true
      try {
        const res = await deleteIntentionStudentImportTaskApi({ taskId: importId.value })
        if (res.code === 200) {
          messageService.success('已取消导入并删除记录')
          router.replace('/import-center/import-intention-student-starter')
          return
        }
        return Promise.reject(new Error(res.message || '删除失败'))
      }
      catch (error) {
        console.error('delete intention student import task failed', error)
        messageService.error('取消导入失败，请稍后重试')
        return Promise.reject(error)
      }
      finally {
        deletingTask.value = false
      }
    },
  })
}

async function loadTaskData() {
  const [detailRes, abnormalRes, normalRes] = await Promise.all([
    getIntentionStudentImportTaskDetailApi({ taskId: importId.value }),
    getIntentionStudentImportTaskRecordListApi({
      queryModel: { taskId: importId.value, type: 0 },
      sortModel: '',
      pageRequestModel: { needTotal: true, pageSize: 1000, pageIndex: 1, skipCount: 0 },
    }),
    getIntentionStudentImportTaskRecordListApi({
      queryModel: { taskId: importId.value, type: 1 },
      sortModel: '',
      pageRequestModel: { needTotal: true, pageSize: 1000, pageIndex: 1, skipCount: 0 },
    }),
  ])

  const detail = detailRes.result || detailRes.data
  const abnormal = abnormalRes.result || abnormalRes.data
  const normal = normalRes.result || normalRes.data

  session.fileName = detail?.fileName || ''
  session.instName = detail?.instName || '总机构'
  session.columns = abnormal?.columns?.length ? abnormal.columns : (normal?.columns || [])
  session.rows = [...(abnormal?.list || []), ...(normal?.list || [])]
  recomputeSummary()
  refreshRowOptionSelections()
}

function buildOptionsByTitle({ defaultFields = [], customFields = [], channels = [], staffs = [] }) {
  const map = {
    手机号归属人: phoneRelationshipOptions,
    性别: sexOptions,
    渠道: channels,
    销售员: staffs,
  }

  const gradeField = defaultFields.find(item => item.fieldKey === '年级')
  if (gradeField?.optionsJson) {
    map.年级 = gradeField.optionsJson.split(',').map(item => item.trim()).filter(Boolean).map(item => ({ label: item, value: item }))
  }

  customFields.forEach((field) => {
    if (field.fieldType === 4 && field.isDisplay) {
      map[field.fieldKey] = (field.optionsJson || '').split(',').map(item => item.trim()).filter(Boolean).map(item => ({ label: item, value: item }))
    }
  })
  return map
}

async function loadOptionSources() {
  const [defaultRes, customRes, channelRes, staffRes] = await Promise.all([
    getStuDefaultFieldApi(),
    getStuCustomFieldApi(),
    getChannelTreeApi(),
    getStaffSummariesApi({
      queryModel: {
        schoolId: '',
        status: 1,
      },
      pageRequestModel: {
        needTotal: true,
        skipCount: 0,
        pageSize: 1000,
        pageIndex: 1,
      },
    }),
  ])

  const channels = []
  ;(channelRes.result || []).forEach((group) => {
    ;(group.channelList || []).forEach((channel) => {
      if (!channel.isDisabled) {
        channels.push({
          label: channel.name,
          value: channel.id,
        })
      }
    })
  })

  const staffs = (staffRes.result?.list || staffRes.data?.list || []).filter(item => item.status === 1).map(item => ({
    label: item.name,
    value: item.id,
  }))

  optionMap.value = buildOptionsByTitle({
    defaultFields: defaultRes.result || [],
    customFields: customRes.result || [],
    channels,
    staffs,
  })
  refreshRowOptionSelections()
}

onMounted(() => {
  taskLoading.value = true
  Promise.all([loadTaskData(), loadOptionSources()]).catch((error) => {
    console.error('load intention student import task failed', error)
    messageService.error('导入任务加载失败')
    router.replace('/import-center/import-intention-student-starter')
  }).finally(() => {
    taskLoading.value = false
  })
})
</script>

<template>
  <div class="import-edit-layout">
    <div class="work-top flex justify-between items-center h-56px bg-#fff">
      <div class="work-top-left flex items-center">
        <div class="import-header-logo" title="导入中心" aria-hidden="true" />
        <span class="text-20px text-#06f font500 ml-16px flex items-center cursor-pointer" @click="handleBack">
          <LeftOutlined class="mt--1px" /> 返回
        </span>
      </div>
      <div class="work-top-right pr-20px text-16px text-#000 font500">
        当前机构：{{ session.instName || '总机构' }}
      </div>
    </div>

    <div class="work-main">
      <div class="work-main-card">
        <div class="title-row">
          <div class="file-title">
            <img src="https://pcsys.admin.ybc365.com/e8183085-4314-4fdf-a9b1-f1934defad7c.png" alt="">
            <span>{{ session.fileName || '正在解析导入文件...' }}</span>
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
            <a-button v-if="hasAbnormalRows" type="primary" ghost @click="handleSave">
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
                <tr v-for="row in displayedRows" :key="row.rowNo">
                  <td class="index-column">{{ row.rowNo }}</td>
                  <td v-for="cell in row.cells" :key="cell.key">
                    <template v-if="activeTab === 'abnormal' && (optionMap[cell.title] || []).length > 0">
                      <a-select
                        :value="cell.selectedId ?? undefined"
                        allow-clear
                        style="width: 180px"
                        placeholder="请选择"
                        @change="value => handleCellChange(row, cell, session.columns.find(col => col.key === cell.key), value)"
                      >
                        <a-select-option
                          v-for="option in optionMap[cell.title] || []"
                          :key="option.value"
                          :value="option.value"
                        >
                          {{ option.label }}
                        </a-select-option>
                      </a-select>
                    </template>
                    <template v-else-if="activeTab === 'abnormal'">
                      <a-input
                        :value="cell.value"
                        :placeholder="session.columns.find(col => col.key === cell.key)?.fieldType === 3 ? '请选择日期格式填写' : '请输入'"
                        style="width: 180px"
                        @input="event => handleCellChange(row, cell, session.columns.find(col => col.key === cell.key), event.target.value)"
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
                    <a-empty />
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
      @cancel="editFormRef?.resetFields?.()"
      @ok="handleConfirmEditModal"
    >
        <a-form ref="editFormRef" layout="vertical" :model="editModalState">
          <a-form-item :name="currentEditingOptions.length > 0 ? 'selectedId' : 'value'" :rules="editFieldRules">
          <template #label>{{ editModalState.title }}</template>
          <template v-if="currentEditingOptions.length > 0">
            <a-select
              v-model:value="editModalState.selectedId"
              show-search
              placeholder="请选择"
              style="width: 100%"
              option-filter-prop="label"
            >
              <a-select-option
                v-for="option in currentEditingOptions"
                :key="option.value"
                :value="option.value"
                :label="option.label"
              >
                {{ option.label }}
              </a-select-option>
            </a-select>
          </template>
          <template v-else-if="currentEditingColumn?.fieldType === 3">
            <a-date-picker
              v-model:value="editModalState.value"
              value-format="YYYY-MM-DD"
              format="YYYY-MM-DD"
              placeholder="请选择日期"
              style="width: 100%"
            />
          </template>
          <template v-else>
            <a-input v-model:value="editModalState.value" placeholder="请输入" />
          </template>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.import-edit-layout {
  min-height: 100vh;
  background: #f7f7fd;
}

.import-header-logo {
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.22) 0%, transparent 42%),
    linear-gradient(145deg, #2b8cff 0%, #0066ff 45%, #0050d8 100%);
  position: relative;
  overflow: hidden;
}

.import-header-logo::before {
  content: '';
  position: absolute;
  left: 12px;
  top: 15px;
  width: 32px;
  height: 26px;
  background-color: rgba(255, 255, 255, 0.94);
  background-image:
    linear-gradient(rgba(0, 102, 255, 0.22), rgba(0, 102, 255, 0.22)),
    linear-gradient(rgba(0, 102, 255, 0.18), rgba(0, 102, 255, 0.18)),
    linear-gradient(rgba(0, 102, 255, 0.14), rgba(0, 102, 255, 0.14));
  background-size: 24px 2px, 18px 2px, 22px 2px;
  background-position: 4px 8px, 4px 14px, 4px 20px;
  background-repeat: no-repeat;
}

.work-main {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

.work-main-card {
  width: 1300px;
  padding: 40px 48px 48px;
  border-radius: 24px;
  background: #fff;
  box-shadow: 0 0 32px rgba(0, 0, 0, 0.08);
}

.task-loading-panel {
  min-height: 360px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
}

.task-loading-title {
  color: #222;
  font-size: 18px;
  font-weight: 600;
}

.task-loading-desc {
  color: #8a8f99;
  font-size: 14px;
}

.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.file-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 600;
  color: #222;
}

.file-title img {
  width: 32px;
  height: 32px;
}

.actions {
  display: flex;
  align-items: center;
}

.tab-row {
  margin-top: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.tabs {
  display: flex;
  gap: 24px;
}

.tab {
  padding-bottom: 2px;
  color: #666;
  cursor: pointer;
  font-size: 16px;
  font-weight: 800;
}

.tab.active {
  color: #06f;
  border-bottom: 3px solid #06f;
}

.table-wrap {
  margin-top: 24px;
  overflow-x: auto;
  overflow-y: hidden;
}

.edit-table {
  width: max-content;
  min-width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.edit-table th,
.edit-table td {
  padding: 14px 12px;
  border-bottom: 1px solid #f0f0f0;
  vertical-align: top;
  white-space: nowrap;
  background: #fff;
  text-align: left;
}

.edit-table th {
  background: #fafafa;
  font-weight: 600;
  color: #222;
}

/* 序号列仅展示数字，与其它列表单项等高时上下居中 */
.edit-table th.index-column,
.edit-table td.index-column {
  vertical-align: middle;
  text-align: center;
}

.edit-table .action-column {
  position: sticky;
  right: 0;
  z-index: 2;
  box-shadow: -8px 0 12px rgba(255, 255, 255, 0.96);
}

.edit-table th.action-column {
  z-index: 3;
  background: #fafafa;
}

.required {
  color: #ff4d4f;
  margin-right: 2px;
}

.error-text {
  margin-top: 4px;
  color: #ff4d4f;
  font-size: 12px;
}

.readonly-cell {
  position: relative;
  display: block;
  width: 100%;
  min-width: 0;
  padding: 6px 28px 6px 6px;
  border: 1px solid transparent;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  box-sizing: border-box;
  margin-left: -6px;
}

.readonly-cell:hover {
  border-color: #d6e4ff;
  background: #f5f9ff;
}

.readonly-cell__text {
  display: block;
  color: #222;
  text-align: left;
}

.readonly-cell__edit {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  color: #1677ff;
  font-size: 12px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.readonly-cell:hover .readonly-cell__edit {
  opacity: 1;
}

.modal-error-text {
  margin-top: 8px;
  color: #ff4d4f;
  font-size: 12px;
}


.empty-table-cell {
  padding: 24px 16px 32px !important;
  vertical-align: middle;
  text-align: center;
  border-bottom: 1px solid #f0f0f0;
  background: #fff;

  :deep(.ant-empty) {
    margin: 0 auto;
  }
}
</style>
