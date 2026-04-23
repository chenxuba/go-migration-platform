<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { DownOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import dayjs from 'dayjs'
import messageService from '@/utils/messageService'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import {
  downloadPendingRenewalStudentExportRecordApi,
  exportPendingRenewalStudentsApi,
  getPendingRenewalStudentExportRecordsApi,
  getPendingRenewalStudentsPagedListApi,
  type PendingRenewalExportConditionItem,
  type PendingRenewalStudentExportRecord,
  type PendingRenewalStudentItem,
  sendPendingRenewalWechatReminderApi,
} from '@/api/edu-center/student-list'
import { useTableColumns } from '@/composables/useTableColumns'
import { useStudentListRefresh } from '@/composables/useStudentListRefresh'
import { Sex, SexLabel } from '@/enums'
import PendingRenewalMessageRecordDrawer from './pending-renewal-message-record-drawer.vue'

const tipsText = '（规则：剩余课时<15 / 剩余天数<15 / 剩余金额<500元）'
const displayArray = ['intentionCourse', 'className', 'classTeacher', 'currentStatus']
const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

type FilterConditionValue = {
  value?: string
}

type FilterConditionItem = {
  label?: string
  values?: FilterConditionValue[]
}

const allFilterRef = ref<{
  clearQuickFilter?: (id?: string | number, type?: string) => void
  getOrderedConditions?: () => FilterConditionItem[]
} | null>(null)
const loading = ref(false)
const dataSource = ref<PendingRenewalStudentItem[]>([])
const classNameOptionsData = ref<Array<{ id: string, value: string }>>([])
const selectedRowKeys = ref<string[]>([])
const selectedRows = ref<PendingRenewalStudentItem[]>([])
const messageRecordOpen = ref(false)
const sendingWechatReminder = ref(false)
const messageRecordDrawerRef = ref<InstanceType<typeof PendingRenewalMessageRecordDrawer> | null>(null)
const exportModalVisible = ref(false)
const exportRecordModalVisible = ref(false)
const exportSubmitting = ref(false)
const exportRecordsLoading = ref(false)
const exportMode = ref('all')
const exportReportType = ref('student')
const exportFileType = ref('excel')
const exportConditionItems = ref<PendingRenewalExportConditionItem[]>([])
const exportModalConditionItems = ref<PendingRenewalExportConditionItem[]>([])
const exportRecords = ref<PendingRenewalStudentExportRecord[]>([])
const summary = ref({
  total: 0,
  studentCount: 0,
})

const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  pageSizeOptions: ['20', '50', '100'],
  showTotal: (total: number) => `共 ${total} 条`,
})

const queryState = ref({
  studentId: undefined as string | undefined,
  productId: undefined as string | undefined,
  classTeacherId: undefined as string | undefined,
  classIds: undefined as string[] | undefined,
  statusList: undefined as number[] | undefined,
})

const allColumns = ref([
  {
    title: '学员/性别',
    dataIndex: 'student',
    key: 'student',
    fixed: 'left',
    width: 180,
    required: true,
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    key: 'phone',
    width: 140,
  },
  {
    title: '当前状态',
    dataIndex: 'status',
    key: 'status',
    width: 120,
  },
  {
    title: '在读课程',
    dataIndex: 'lessonName',
    key: 'lessonName',
    width: 180,
  },
  {
    title: '班主任',
    dataIndex: 'classTeacherList',
    key: 'classTeacherList',
    width: 180,
  },
  {
    title: '剩余数量',
    dataIndex: 'remaining',
    key: 'remaining',
    width: 180,
  },
  {
    title: '到期时间',
    dataIndex: 'expireTime',
    key: 'expireTime',
    width: 140,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'pending-renewal-student-list',
  allColumns,
  excludeKeys: ['action'],
})

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: Array<string | number>, rows: PendingRenewalStudentItem[]) => {
    selectedRowKeys.value = keys.map(item => String(item))
    selectedRows.value = rows
  },
}))

const selectedCount = computed(() => selectedRowKeys.value.length)
const exportPreviewColumns = [
  { title: '学员姓名', dataIndex: 'studentName', key: 'studentName' },
  { title: '性别', dataIndex: 'sex', key: 'sex' },
  { title: '学员手机号', dataIndex: 'phone', key: 'phone' },
  { title: '当前状态', dataIndex: 'status', key: 'status' },
  { title: '在读课程', dataIndex: 'lessonName', key: 'lessonName' },
  { title: '班主任', dataIndex: 'classTeacher', key: 'classTeacher' },
  { title: '收费模式', dataIndex: 'chargingMode', key: 'chargingMode' },
  { title: '剩余数量', dataIndex: 'remainingQuantity', key: 'remainingQuantity' },
  { title: '剩余学费金额', dataIndex: 'remainingTuition', key: 'remainingTuition' },
  { title: '到期时间', dataIndex: 'expireTime', key: 'expireTime' },
]
const exportPreviewRows = [
  {
    studentName: '王小明',
    sex: '男',
    phone: '18818888888',
    status: '正常',
    lessonName: '篮球课',
    classTeacher: '王老师',
    chargingMode: '按课时',
    remainingQuantity: '10',
    remainingTuition: '10',
    expireTime: '2022-05-18 00:00:00',
  },
]
const exportFieldCount = computed(() => exportPreviewColumns.length)
const exportQuerySummary = computed(() => {
  if (exportModalConditionItems.value.length === 0)
    return ['全部导出']
  return exportModalConditionItems.value.map(item => `${item.label}：${item.value}`)
})

function resetQueryState() {
  queryState.value.studentId = undefined
  queryState.value.productId = undefined
  queryState.value.classTeacherId = undefined
  queryState.value.classIds = undefined
  queryState.value.statusList = undefined
}

function normalizeStringValue(value: unknown) {
  if (value === undefined || value === null)
    return undefined
  const text = String(value).trim()
  return text || undefined
}

function normalizeStringArray(value: unknown) {
  if (!Array.isArray(value))
    return undefined
  const list = value
    .map(item => String(item ?? '').trim())
    .filter(Boolean)
  return list.length ? list : undefined
}

function normalizeStatusList(value: unknown) {
  if (!Array.isArray(value))
    return undefined
  const list = value
    .map(item => Number(item))
    .filter(item => Number.isFinite(item))
  return list.length ? list : undefined
}

async function loadClassNameOptions() {
  try {
    const lessonIds = queryState.value.productId ? [queryState.value.productId] : undefined
    const res = await pageGroupClassesApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 200,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        lessonIds,
      },
    })

    if (res.code !== 200) {
      throw new Error(res.message || '获取班级筛选项失败')
    }

    const list = Array.isArray(res.result?.list) ? res.result.list : []
    const optionMap = new Map<string, { id: string, value: string }>()
    list.forEach((item) => {
      const id = String(item?.id ?? '').trim()
      const value = String(item?.name ?? '').trim()
      if (!id || !value || optionMap.has(id))
        return
      optionMap.set(id, { id, value })
    })
    classNameOptionsData.value = [...optionMap.values()]
  }
  catch (error) {
    console.error('加载待续费班级筛选项失败:', error)
    classNameOptionsData.value = []
  }
}

function formatNumber(value?: number) {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return '0'
  return Number.isInteger(num) ? String(num) : num.toFixed(2).replace(/\.?0+$/, '')
}

function formatMoney(value?: number) {
  const num = Number(value || 0)
  return num.toFixed(2)
}

function isAmountMode(mode?: number) {
  const value = Number(mode || 0)
  return value === 3 || value === 4
}

function getRemainingUnit(mode?: number) {
  const value = Number(mode || 0)
  if (value === 2)
    return '天'
  if (isAmountMode(value))
    return '元'
  return '课时'
}

function getGenderText(sex?: number) {
  const value = Number.isFinite(Number(sex)) ? Number(sex) : Sex.Unknown
  return SexLabel[value as Sex] || SexLabel[Sex.Unknown]
}

function getStatusInfo(status?: number) {
  const statusValue = Number(status || 0)
  const map: Record<number, { text: string, className: string }> = {
    1: { text: '正常', className: 'text-#0c3 bg-#e6ffec' },
    2: { text: '已停课', className: 'text-#f90 bg-#fff5e6' },
    3: { text: '已结课', className: 'text-#888 bg-#f5f5f5' },
  }
  return map[statusValue] || { text: '未知', className: 'text-#888 bg-#f5f5f5' }
}

function formatExpireDate(record: PendingRenewalStudentItem) {
  if (!record.enableExpireTime || !record.expireTime)
    return '-'
  const date = dayjs(record.expireTime)
  if (!date.isValid() || date.year() <= 1)
    return '-'
  return date.format('YYYY-MM-DD')
}

function getClassTeacherText(record: PendingRenewalStudentItem) {
  const teacherNames = Array.isArray(record.classTeacherList)
    ? record.classTeacherList
      .map(item => String(item?.name ?? '').trim())
      .filter(Boolean)
    : []
  return teacherNames.length ? teacherNames.join('、') : '-'
}

function getRemainingText(record: PendingRenewalStudentItem) {
  if (isAmountMode(record.lessonChargingMode))
      return `¥ ${formatMoney(record.tuition || record.leftQuantity)}`
  const total = Number(record.leftQuantity || 0) + Number(record.leftFreeQuantity || 0)
  return `${formatNumber(total)}${getRemainingUnit(record.lessonChargingMode)}`
}

function getRemainingSubText(record: PendingRenewalStudentItem) {
  if (isAmountMode(record.lessonChargingMode))
    return ''
  const freeQuantity = Number(record.leftFreeQuantity || 0)
  if (freeQuantity > 0) {
    return `正课${formatNumber(record.leftQuantity)}${getRemainingUnit(record.lessonChargingMode)} + 赠送${formatNumber(freeQuantity)}${getRemainingUnit(record.lessonChargingMode)}`
  }
  return ''
}

function buildExportQueryModel() {
  return Object.fromEntries(
    Object.entries(queryState.value).filter(([, value]) => value !== undefined && value !== null && (!Array.isArray(value) || value.length > 0)),
  )
}

function syncExportConditions() {
  const orderedConditions = allFilterRef.value?.getOrderedConditions?.() || []
  const mappedConditions = orderedConditions
    .map((item) => {
      const label = String(item?.label || '').trim()
      if (!label)
        return null
      const values = Array.isArray(item?.values) ? item.values : []
      const valueText = values.length > 0
        ? values.map(valueItem => String(valueItem?.value || '').trim().replace(' 至 ', ' ~ ')).filter(Boolean).join('、')
        : '全部'
      return {
        label,
        value: valueText || '全部',
      }
    })
    .filter((item): item is PendingRenewalExportConditionItem => Boolean(item))

  exportModalConditionItems.value = [...mappedConditions]
  exportConditionItems.value = [...mappedConditions]
}

function getExportRecordDisplayConditions(record: PendingRenewalStudentExportRecord) {
  const conditions = Array.isArray(record?.queryConditions) ? record.queryConditions : []
  if (!conditions.length) {
    return [{ label: '', value: '全部导出' }]
  }
  return conditions.map(item => ({
    label: String(item?.label || '').trim(),
    value: String(item?.value || '').trim() || '全部',
  }))
}

async function openExportModal() {
  syncExportConditions()
  exportModalVisible.value = true
}

async function openExportRecordModal() {
  syncExportConditions()
  exportRecordModalVisible.value = true
  await loadExportRecords()
}

function triggerBlobDownload(response: any) {
  const blob = new Blob([response.data], { type: response.headers['content-type'] || 'application/octet-stream' })
  const disposition = response.headers['content-disposition'] || ''
  const matched = disposition.match(/filename\*=UTF-8''([^;]+)/i)
  const fileName = matched ? decodeURIComponent(matched[1]) : `待续费学员批量导出-${dayjs().format('YYYYMMDDHHmmss')}.xlsx`
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

async function loadExportRecords() {
  exportRecordsLoading.value = true
  try {
    const res = await getPendingRenewalStudentExportRecordsApi()
    if (res.code !== 200) {
      throw new Error(res.message || '获取导出记录失败')
    }
    exportRecords.value = Array.isArray(res.result) ? res.result : []
  }
  catch (error: any) {
    console.error('load pending renewal export records failed', error)
    messageService.error(error?.message || '获取导出记录失败')
  }
  finally {
    exportRecordsLoading.value = false
  }
}

async function downloadExportRecord(record: PendingRenewalStudentExportRecord) {
  try {
    const response = await downloadPendingRenewalStudentExportRecordApi(record.id)
    triggerBlobDownload(response)
  }
  catch (error: any) {
    console.error('download pending renewal export record failed', error)
    messageService.error(error?.message || '下载失败，请稍后重试')
  }
}

async function handleViewExportRecord() {
  exportModalVisible.value = false
  await openExportRecordModal()
}

async function handleSubmitExport() {
  if (summary.value.total === 0 || dataSource.value.length === 0) {
    messageService.error('没有符合条件的待续费学员可以导出')
    return
  }
  exportSubmitting.value = true
  try {
    syncExportConditions()
    const res = await exportPendingRenewalStudentsApi({
      queryModel: buildExportQueryModel(),
      queryConditions: exportConditionItems.value,
    })
    const record = res.result
    if (!record?.id) {
      throw new Error(res.message || '导出失败')
    }
    const response = await downloadPendingRenewalStudentExportRecordApi(record.id)
    triggerBlobDownload(response)
    exportModalVisible.value = false
    await openExportRecordModal()
  }
  catch (error: any) {
    console.error('export pending renewal students failed', error)
    messageService.error(error?.message || '导出失败，请稍后重试')
  }
  finally {
    exportSubmitting.value = false
  }
}

async function handleExportAction({ key }: { key: string | number }) {
  const actionKey = String(key)
  if (actionKey === '1') {
    await openExportModal()
    return
  }
  if (actionKey === '2')
    await openExportRecordModal()
}

function handleMessageRecord() {
  messageRecordOpen.value = true
}

async function handleWechatRemind() {
  if (!selectedCount.value) {
    messageService.warning('请先选择待发送的学员')
    return
  }
  if (sendingWechatReminder.value)
    return
  sendingWechatReminder.value = true
  try {
    const res = await sendPendingRenewalWechatReminderApi({
      tuitionAccountIds: selectedRowKeys.value,
    })
    if (res.code !== 200) {
      throw new Error(res.message || '发送续费提醒失败')
    }
    const result = res.result
    const successCount = Number(result?.successCount || 0)
    const skippedCount = Number(result?.skippedCount || 0)
    const failedCount = Number(result?.failedCount || 0)

    if (successCount > 0) {
      messageService.success(`已发送 ${successCount} 条微信提醒${skippedCount > 0 ? `，未关注跳过 ${skippedCount} 条` : ''}${failedCount > 0 ? `，失败 ${failedCount} 条` : ''}`)
    }
    else if (skippedCount > 0 || failedCount > 0) {
      messageService.warning(`本次未成功发送${skippedCount > 0 ? `，未关注跳过 ${skippedCount} 条` : ''}${failedCount > 0 ? `，失败 ${failedCount} 条` : ''}`)
    }
    else {
      messageService.info('本次没有可发送的续费提醒')
    }

    selectedRowKeys.value = []
    selectedRows.value = []
    messageRecordOpen.value = true
    await nextTick()
    await messageRecordDrawerRef.value?.getList?.()
  }
  catch (error: any) {
    console.error('send pending renewal wechat reminder failed', error)
    messageService.error(error?.message || '发送续费提醒失败')
  }
  finally {
    sendingWechatReminder.value = false
  }
}

function handleSmsRemind() {
  if (!selectedCount.value) {
    messageService.warning('请先选择待发送的学员')
    return
  }
  messageService.info('短信提醒功能待接入')
}

async function getList(id?: string | number, type?: string) {
  loading.value = true
  try {
    const queryModel = Object.fromEntries(
      Object.entries(queryState.value).filter(([, value]) => value !== undefined),
    )

    const res = await getPendingRenewalStudentsPagedListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: 0,
      },
      queryModel,
      sortModel: {
        expriedTime: 0,
      },
    })

    if (res.code !== 200) {
      throw new Error(res.message || '获取待续费学员列表失败')
    }

    const result = res.result || {}
    dataSource.value = Array.isArray(result.list) ? result.list : []
    pagination.value.total = Number(result.total || 0)
    summary.value.total = Number(result.total || 0)
    summary.value.studentCount = Number(result.studentCount || 0)
    allFilterRef.value?.clearQuickFilter?.(id, type)
  }
  catch (error: any) {
    console.error('获取待续费学员列表失败:', error)
    messageService.error(error?.message || '获取待续费学员列表失败')
  }
  finally {
    loading.value = false
  }
}

function applyFilterChange(
  updates: Partial<typeof queryState.value>,
  id?: string | number,
  type?: string,
  options?: { reloadClassOptions?: boolean },
) {
  Object.assign(queryState.value, updates)
  pagination.value.current = 1
  selectedRowKeys.value = []
  selectedRows.value = []
  if (options?.reloadClassOptions)
    void loadClassNameOptions()
  void getList(id, type)
}

const filterUpdateHandlers = computed(() => ({
  'update:stuPhoneSearchFilter': (val: unknown, isClearAll?: boolean, id?: string | number, type?: string) => {
    if (isClearAll) {
      resetQueryState()
      void loadClassNameOptions()
      void getList(id, type)
      return
    }
    applyFilterChange({ studentId: normalizeStringValue(val) }, id, type)
  },
  'update:intentionCourseFilter': (val: unknown, isClearAll?: boolean, id?: string | number, type?: string) => {
    if (isClearAll) {
      resetQueryState()
      void loadClassNameOptions()
      void getList(id, type)
      return
    }
    applyFilterChange({ productId: normalizeStringValue(val) }, id, type, { reloadClassOptions: true })
  },
  'update:classTeacherFilter': (val: unknown, isClearAll?: boolean, id?: string | number, type?: string) => {
    if (isClearAll) {
      resetQueryState()
      void loadClassNameOptions()
      void getList(id, type)
      return
    }
    applyFilterChange({ classTeacherId: normalizeStringValue(val) }, id, type)
  },
  'update:classNameFilter': (val: unknown, isClearAll?: boolean, id?: string | number, type?: string) => {
    if (isClearAll) {
      resetQueryState()
      void loadClassNameOptions()
      void getList(id, type)
      return
    }
    applyFilterChange({ classIds: normalizeStringArray(val) }, id, type)
  },
  'update:currentStatusFilter': (val: unknown, isClearAll?: boolean, id?: string | number, type?: string) => {
    if (isClearAll) {
      resetQueryState()
      void loadClassNameOptions()
      void getList(id, type)
      return
    }
    applyFilterChange({ statusList: normalizeStatusList(val) }, id, type)
  },
}))

function handleTableChange(paginationInfo: any) {
  pagination.value.current = Number(paginationInfo?.current || 1)
  pagination.value.pageSize = Number(paginationInfo?.pageSize || pagination.value.pageSize)
  void getList()
}

onMounted(async () => {
  await loadClassNameOptions()
  await getList()
})

useStudentListRefresh(getList)

defineExpose({
  getList,
})
</script>

<template>
  <div>
    <div class="filter-wrap mt-2 bg-white pl-3 pr-3 rounded-4">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="true"
        :class-name-options-data="classNameOptionsData"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共{{ summary.studentCount }}名学员，{{ summary.total }}条待续费记录
            <span class="text-#0066ff">{{ tipsText }}</span>
          </div>
          <div class="edit flex">
            <a-button class="mr-2" @click="handleMessageRecord">
              消息记录
            </a-button>
            <a-dropdown class="mr-2" overlay-class-name="student-export-dropdown">
              <template #overlay>
                <a-menu @click="handleExportAction">
                  <a-menu-item key="1">
                    批量导出
                  </a-menu-item>
                  <a-menu-item key="2">
                    导出记录
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导出数据
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1" @click="handleWechatRemind">
                    微信提醒
                  </a-menu-item>
                  <a-menu-item key="2" @click="handleSmsRemind">
                    短信提醒
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button :loading="sendingWechatReminder">
                批量发送续费提醒{{ selectedCount > 0 ? `(${selectedCount})` : '' }}
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <customize-code
              v-model:checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length"
              :num="selectedValues.length"
            />
          </div>
        </div>

        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :loading="loading"
            :pagination="pagination"
            :columns="filteredColumns"
            :row-selection="rowSelection"
            :scroll="{ x: totalWidth }"
            row-key="tuitionAccountId"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'student'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :gender="getGenderText(record.sex)"
                  :avatar-url="record.avatar || ''"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>

              <template v-if="column.key === 'phone'">
                <div class="text-#222">
                  {{ record.phone || '-' }}
                </div>
              </template>

              <template v-if="column.key === 'status'">
                <span :class="`${getStatusInfo(record.status).className} rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2`">
                  {{ getStatusInfo(record.status).text }}
                </span>
              </template>

              <template v-if="column.key === 'lessonName'">
                <div class="text-#222">
                  {{ record.lessonName || '-' }}
                </div>
              </template>

              <template v-if="column.key === 'classTeacherList'">
                <div class="text-#222">
                  {{ getClassTeacherText(record) }}
                </div>
              </template>

              <template v-if="column.key === 'remaining'">
                <div class="text-#222">
                  {{ getRemainingText(record) }}
                </div>
                <div v-if="getRemainingSubText(record)" class="text-3 text-#888">
                  {{ getRemainingSubText(record) }}
                </div>
              </template>

              <template v-if="column.key === 'expireTime'">
                {{ formatExpireDate(record) }}
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <pending-renewal-message-record-drawer
      ref="messageRecordDrawerRef"
      v-model:open="messageRecordOpen"
    />

    <a-modal
      v-model:open="exportModalVisible"
      title="批量导出"
      :footer="null"
      :width="820"
      class="student-export-modal"
      destroy-on-close
    >
      <div class="export-tip-bar">
        <InfoCircleOutlined class="export-tip-icon" />
        <span>当前列表最多支持导出 10000 条数据。若超出，请前往【数据中心-报表管理-明细表】导出</span>
      </div>

      <div class="export-modal-content">
        <div class="export-row">
          <div class="export-label">
            查询条件：
          </div>
          <div class="export-query-box">
            <div v-for="item in exportQuerySummary" :key="item" class="export-query-line">
              {{ item }}
            </div>
          </div>
        </div>

        <div class="export-row export-row--compact">
          <div class="export-label">
            导出方式：
          </div>
          <a-radio-group v-model:value="exportMode" class="custom-radio export-radio-group">
            <a-radio value="all">
              全部导出
            </a-radio>
          </a-radio-group>
        </div>

        <div class="export-row export-row--compact">
          <div class="export-label">
            报表类型：
          </div>
          <a-radio-group v-model:value="exportReportType" class="custom-radio export-radio-group">
            <a-radio value="student">
              学员维度
            </a-radio>
          </a-radio-group>
        </div>

        <div class="export-row export-row--stacked">
          <div class="export-label">
            导出范例：
          </div>
          <div class="export-preview-title">
            共{{ exportFieldCount }}个字段
          </div>
          <div class="export-preview-card">
            <div class="export-preview-scroll">
              <a-table
                :data-source="exportPreviewRows"
                :columns="exportPreviewColumns"
                :pagination="false"
                size="small"
                :scroll="{ x: 1400 }"
                row-key="studentName"
              />
            </div>
          </div>
        </div>

        <div class="export-row export-row--compact">
          <div class="export-label">
            生成类型：
          </div>
          <a-radio-group v-model:value="exportFileType" class="custom-radio export-radio-group">
            <a-radio value="excel">
              EXCEL格式文件
            </a-radio>
          </a-radio-group>
        </div>
      </div>

      <div class="export-modal-footer">
        <a-button @click="handleViewExportRecord">
          查看导出记录
        </a-button>
        <a-button type="primary" class="ml-3" :loading="exportSubmitting" @click="handleSubmitExport">
          导出
        </a-button>
      </div>
    </a-modal>

    <a-modal
      v-model:open="exportRecordModalVisible"
      title="导出记录"
      :footer="null"
      :width="800"
      class="student-export-record-modal"
      destroy-on-close
    >
      <a-spin :spinning="exportRecordsLoading">
        <div v-if="exportRecords.length > 0" class="export-record-list">
          <div v-for="record in exportRecords" :key="record.id" class="export-record-card">
            <div class="export-record-header">
              <div class="export-record-meta">
                <span>报表生成时间：{{ record.createdTime ? dayjs(record.createdTime).format('YYYY-MM-DD HH:mm:ss') : '-' }}</span>
                <span class="ml-6">导出人：{{ record.exporterName || '-' }}</span>
              </div>
              <a-button @click="downloadExportRecord(record)">
                下载
              </a-button>
            </div>

            <div class="export-record-body">
              <div class="export-record-top">
                <div class="export-record-title">
                  查询条件
                </div>
                <div class="export-record-expire">
                  请在一周内下载，过期将失效
                </div>
              </div>
              <div class="export-record-grid">
                <div v-for="item in getExportRecordDisplayConditions(record)" :key="`${record.id}-${item.label}-${item.value}`" class="export-record-item">
                  <span v-if="item.label" class="export-record-item-label">{{ item.label }}：</span>
                  <span>{{ item.value }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-else-if="!exportRecordsLoading" class="export-record-empty">
          <a-empty :image="simpleImage" description="暂无数据" />
        </div>
      </a-spin>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

:deep(.student-export-modal .ant-modal-body),
:deep(.student-export-record-modal .ant-modal-body) {
  padding-top: 0;
}

.export-tip-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 -24px;
  padding: 12px 20px;
  background: #eaf3ff;
  color: #1668dc;
  font-size: 15px;
  line-height: 22px;
}

.export-tip-icon {
  flex-shrink: 0;
  font-size: 16px;
}

.export-modal-content {
  padding-top: 22px;
}

.export-row {
  display: flex;
  align-items: center;
  margin-bottom: 18px;
}

.export-row--compact {
  align-items: center;
  margin-bottom: 16px;
}

.export-row--stacked {
  display: grid;
  grid-template-columns: 88px minmax(0, 1fr);
  row-gap: 12px;
  align-items: start;
}

.export-label {
  flex-shrink: 0;
  width: 88px;
  color: #595959;
  font-size: 15px;
  line-height: 22px;
}

.export-query-box {
  flex: 1;
  min-height: 56px;
  padding: 16px 18px;
  border-radius: 12px;
  background: #f5f7fb;
  color: #262626;
  font-size: 15px;
  line-height: 24px;
}

.export-query-line + .export-query-line {
  margin-top: 6px;
}

.export-radio-group {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}

.export-radio-group :deep(.ant-radio-wrapper) {
  margin-right: 24px;
  color: #262626;
  font-size: 15px;
  line-height: 22px;
}

.export-preview-title {
  flex: 1;
  color: #262626;
  font-size: 15px;
  line-height: 22px;
}

.export-preview-card {
  flex: 1;
  overflow: hidden;
  border: 1px solid #edf0f5;
  border-radius: 12px;
}

.export-row--stacked .export-preview-card {
  grid-column: 2;
}

.export-preview-scroll {
  overflow-x: auto;
}

.export-modal-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.export-record-list {
  max-height: 520px;
  overflow-y: auto;
  padding-right: 4px;
}

.export-record-empty {
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.export-record-card {
  border: 1px solid #edf0f5;
  border-radius: 12px;
  overflow: hidden;
  background: #fff;
  margin-bottom: 16px;
}

.export-record-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid #edf0f5;
}

.export-record-meta {
  color: #262626;
  font-size: 15px;
  line-height: 24px;
}

.export-record-body {
  padding: 18px 24px 20px;
}

.export-record-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.export-record-title {
  color: #262626;
  font-size: 15px;
  font-weight: 600;
}

.export-record-expire {
  color: #1668dc;
  font-size: 14px;
}

.export-record-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 40px;
}

.export-record-item {
  color: #262626;
  font-size: 15px;
  line-height: 24px;
}

.export-record-item-label {
  color: #595959;
}

.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}

:deep(.student-export-dropdown .ant-dropdown-menu) {
  min-width: 156px;
  padding: 8px 0;
  border-radius: 14px;
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.12);
}

:deep(.student-export-dropdown .ant-dropdown-menu-item) {
  height: 34px;
  padding: 0 16px;
  color: #262626;
  font-size: 14px;
  line-height: 34px;
}

:deep(.student-export-dropdown .ant-dropdown-menu-item:hover) {
  background: #f5f8ff;
}
</style>
