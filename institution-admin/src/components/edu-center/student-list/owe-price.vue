<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { DownOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { Empty, Modal } from 'ant-design-vue'
import type { TableColumnsType } from 'ant-design-vue'
import dayjs from 'dayjs'
import { setBadDebtApi } from '@/api/finance-center/order-manage'
import {
  createStudentLessonArrearExportRecordApi,
  createStudentRegistrationArrearExportRecordApi,
  downloadStudentLessonArrearExportRecordApi,
  downloadStudentRegistrationArrearExportRecordApi,
  getStudentLessonArrearExportRecordsApi,
  getStudentLessonArrearPagedListApi,
  getStudentLessonArrearStatisticsApi,
  getStudentRegistrationArrearExportRecordsApi,
  getStudentRegistrationArrearPagedListApi,
  getStudentRegistrationArrearStatisticsApi,
  type StudentArrearExportConditionItem,
  type StudentArrearExportRecord,
  type StudentLessonArrearItem,
  type StudentRegistrationArrearItem,
} from '@/api/edu-center/student-list'
import { useStudentListRefresh } from '@/composables/useStudentListRefresh'
import messageService from '@/utils/messageService'

type ArrearTabKey = 'registration' | 'lesson'
type ArrearRecord = StudentRegistrationArrearItem | StudentLessonArrearItem

type FilterConditionValue = {
  value?: string
}

type FilterConditionItem = {
  label?: string
  values?: FilterConditionValue[]
}

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE
const activeTab = ref<ArrearTabKey>('registration')
const loading = ref(false)
const exporting = ref(false)
const batchOperating = ref(false)
const selectedRowKeys = ref<Array<string>>([])
const exportModalVisible = ref(false)
const exportRecordModalVisible = ref(false)
const exportSubmitting = ref(false)
const exportRecordsLoading = ref(false)
const exportMode = ref('all')
const exportReportType = ref<ArrearTabKey>('registration')
const exportFileType = ref('excel')
const exportConditionItems = ref<StudentArrearExportConditionItem[]>([])
const exportModalConditionItems = ref<StudentArrearExportConditionItem[]>([])
const exportRecords = ref<StudentArrearExportRecord[]>([])
const allFilterRef = ref<{
  clearQuickFilter?: (id?: string | number, type?: string) => void
  getOrderedConditions?: () => FilterConditionItem[]
} | null>(null)
const registrationDisplayArray = ['orderNumber', 'intentionCourse', 'createTime']
const lessonDisplayArray = ['intentionCourse']

const registrationList = ref<StudentRegistrationArrearItem[]>([])
const registrationSummary = ref({
  total: 0,
  totalArrearAmount: 0,
})
const registrationPagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
})
const registrationFilters = ref({
  orderNumber: '',
  lessonId: undefined as string | undefined,
  studentId: undefined as string | undefined,
  createdTimeBegin: '',
  createdTimeEnd: '',
})

const lessonList = ref<StudentLessonArrearItem[]>([])
const lessonSummary = ref({
  total: 0,
  totalArrearAmount: 0,
  totalArrearTime: 0,
})
const lessonPagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
})
const lessonFilters = ref({
  lessonId: undefined as string | undefined,
  studentId: undefined as string | undefined,
})
const initializedDefaultTab = ref(false)

const registrationColumns: TableColumnsType<StudentRegistrationArrearItem> = [
  { title: '学员/性别', dataIndex: 'student', key: 'student', width: 220 },
  { title: '联系电话', dataIndex: 'phone', key: 'phone', width: 160 },
  { title: '欠费金额', dataIndex: 'arrearAmount', key: 'arrearAmount', width: 140 },
  { title: '订单总金额', dataIndex: 'orderAmount', key: 'orderAmount', width: 140 },
  { title: '已支付金额', dataIndex: 'paidAmount', key: 'paidAmount', width: 140 },
  { title: '原订单号', dataIndex: 'orderNumber', key: 'orderNumber', width: 220 },
  { title: '商品名称', dataIndex: 'productName', key: 'productName', width: 220 },
  { title: '创建时间', dataIndex: 'createdTime', key: 'createdTime', width: 180 },
  { title: '操作', dataIndex: 'action', key: 'action', width: 100, fixed: 'right' as const },
]

const lessonColumns: TableColumnsType<StudentLessonArrearItem> = [
  { title: '学员/性别', dataIndex: 'student', key: 'student', width: 220 },
  { title: '联系电话', dataIndex: 'phone', key: 'phone', width: 160 },
  { title: '课程商品名称', dataIndex: 'lessonName', key: 'lessonName', width: 220 },
  { title: '欠费项', dataIndex: 'arrearItem', key: 'arrearItem', width: 180 },
  { title: '拖欠记录', dataIndex: 'recordCount', key: 'recordCount', width: 120 },
  { title: '操作', dataIndex: 'action', key: 'action', width: 100, fixed: 'right' as const },
]

const registrationExportPreviewColumns = [
  { title: '学员姓名', dataIndex: 'studentName', key: 'studentName' },
  { title: '性别', dataIndex: 'sex', key: 'sex' },
  { title: '联系电话', dataIndex: 'phone', key: 'phone' },
  { title: '欠费金额（元）', dataIndex: 'arrearAmount', key: 'arrearAmount' },
  { title: '订单总金额（元）', dataIndex: 'orderAmount', key: 'orderAmount' },
  { title: '已支付金额（元）', dataIndex: 'paidAmount', key: 'paidAmount' },
  { title: '原订单号', dataIndex: 'orderNumber', key: 'orderNumber' },
  { title: '商品名称', dataIndex: 'productName', key: 'productName' },
  { title: '创建时间', dataIndex: 'createdTime', key: 'createdTime' },
]

const lessonExportPreviewColumns = [
  { title: '学员姓名', dataIndex: 'studentName', key: 'studentName' },
  { title: '性别', dataIndex: 'sex', key: 'sex' },
  { title: '联系电话', dataIndex: 'phone', key: 'phone' },
  { title: '课程商品名称', dataIndex: 'lessonName', key: 'lessonName' },
  { title: '欠费项', dataIndex: 'arrearItem', key: 'arrearItem' },
  { title: '欠费数值', dataIndex: 'arrearValue', key: 'arrearValue' },
  { title: '欠费单位', dataIndex: 'arrearUnit', key: 'arrearUnit' },
  { title: '拖欠记录（条）', dataIndex: 'recordCount', key: 'recordCount' },
]

const registrationExportPreviewRows = [
  {
    studentName: '王小明',
    sex: '男',
    phone: '18818888888',
    arrearAmount: '300.00',
    orderAmount: '1000.00',
    paidAmount: '700.00',
    orderNumber: 'SO202604230001',
    productName: '篮球课',
    createdTime: '2026-04-23 13:02:02',
  },
]

const lessonExportPreviewRows = [
  {
    studentName: '王小明',
    sex: '男',
    phone: '18818888888',
    lessonName: '篮球课',
    arrearItem: '2课时',
    arrearValue: '2',
    arrearUnit: '课时',
    recordCount: '1',
  },
]

const currentDisplayArray = computed(() => (activeTab.value === 'registration' ? registrationDisplayArray : lessonDisplayArray))
const activeColumns = computed(() => (activeTab.value === 'registration' ? registrationColumns : lessonColumns))
const activeDataSource = computed(() => (activeTab.value === 'registration' ? registrationList.value : lessonList.value))
const activePagination = computed(() => (activeTab.value === 'registration' ? registrationPagination.value : lessonPagination.value))
const totalWidth = computed(() => activeColumns.value.reduce((sum, item) => sum + Number(item.width || 0), 0))
const selectedCount = computed(() => selectedRowKeys.value.length)
const currentSummaryText = computed(() => {
  if (activeTab.value === 'registration')
    return `当前共 ${registrationSummary.value.total} 条数据，欠费金额 ${formatCurrency(registrationSummary.value.totalArrearAmount)}`
  return `当前共 ${lessonSummary.value.total} 条数据，欠费金额 ${formatCurrency(lessonSummary.value.totalArrearAmount)}，欠课时 ${formatLessonArrearTime(lessonSummary.value.totalArrearTime)}`
})
const currentExportName = computed(() => (activeTab.value === 'registration' ? '报名欠费' : '课消欠费'))
const currentExportPreviewColumns = computed(() => (activeTab.value === 'registration' ? registrationExportPreviewColumns : lessonExportPreviewColumns))
const currentExportPreviewRows = computed(() => (activeTab.value === 'registration' ? registrationExportPreviewRows : lessonExportPreviewRows))
const exportFieldCount = computed(() => currentExportPreviewColumns.value.length)
const exportQuerySummary = computed(() => {
  if (exportModalConditionItems.value.length === 0)
    return ['全部导出']
  return exportModalConditionItems.value.map(item => `${item.label}：${item.value}`)
})

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: Array<string | number>) => {
    selectedRowKeys.value = keys.map(key => String(key))
  },
}))

function genderText(value?: number) {
  if (Number(value || 0) === 2)
    return '女'
  if (Number(value || 0) === 1)
    return '男'
  return '未知'
}

function formatCurrency(value?: number) {
  const num = Number(value || 0)
  return `¥ ${num.toFixed(2)}`
}

function formatDateTime(value?: string) {
  if (!value)
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : value
}

function formatLessonArrearTime(value?: number) {
  const num = Number(value || 0)
  if (!Number.isFinite(num) || num <= 0)
    return '0课时'
  const text = Number.isInteger(num) ? String(num) : num.toFixed(2).replace(/\.?0+$/, '')
  return `${text}课时`
}

function formatLessonArrearItem(record: StudentLessonArrearItem) {
  const total = Number(record.beInArrearsTotal || 0)
  const text = Number.isInteger(total) ? String(total) : total.toFixed(2).replace(/\.?0+$/, '')
  if (Number(record.lessonChargingMode || 0) === 3)
    return `${formatCurrency(total)}\n欠费`
  if (Number(record.lessonChargingMode || 0) === 2)
    return `${text}分钟\n欠课时`
  return `${text}课时\n欠课时`
}

function formatLessonArrearItemParts(record: Partial<StudentLessonArrearItem>) {
  const [valueText, labelText] = formatLessonArrearItem({
    studentId: '',
    studentName: '',
    lessonId: '',
    lessonName: '',
    tuitionAccountId: '',
    lessonChargingMode: Number(record.lessonChargingMode || 0),
    beInArrearsTotal: Number(record.beInArrearsTotal || 0),
    recordCount: Number(record.recordCount || 0),
    avatar: '',
    phone: '',
  }).split('\n')
  return {
    valueText: valueText || '-',
    labelText: labelText || '',
  }
}

function normalizeFilterValue(value: unknown) {
  if (value === undefined || value === null)
    return undefined
  const text = String(value).trim()
  return text || undefined
}

function applyFilterChange(payload: {
  registration?: Partial<typeof registrationFilters.value>
  lesson?: Partial<typeof lessonFilters.value>
}, id?: string | number, type?: string) {
  if (payload.registration)
    Object.assign(registrationFilters.value, payload.registration)
  if (payload.lesson)
    Object.assign(lessonFilters.value, payload.lesson)
  if (activeTab.value === 'registration')
    registrationPagination.value.current = 1
  else
    lessonPagination.value.current = 1
  void getList(id, type)
}

const filterUpdateHandlers = computed(() => ({
  'update:orderNumberFilter': (val: unknown, _isClearAll?: boolean, id?: string | number, type?: string) => {
    applyFilterChange({
      registration: {
        orderNumber: normalizeFilterValue(val) || '',
      },
    }, id, type)
  },
  'update:intentionCourseFilter': (val: unknown, _isClearAll?: boolean, id?: string | number, type?: string) => {
    const lessonId = normalizeFilterValue(val)
    applyFilterChange({
      registration: { lessonId },
      lesson: { lessonId },
    }, id, type)
  },
  'update:createTimeFilter': (val: unknown, _isClearAll?: boolean, id?: string | number, type?: string) => {
    const range = Array.isArray(val) ? val.map(item => String(item || '').trim()).filter(Boolean) : []
    applyFilterChange({
      registration: {
        createdTimeBegin: range[0] || '',
        createdTimeEnd: range[1] || '',
      },
    }, id, type)
  },
  'update:stuPhoneSearchFilter': (val: unknown, _isClearAll?: boolean, id?: string | number, type?: string) => {
    const studentId = normalizeFilterValue(val)
    applyFilterChange({
      registration: { studentId },
      lesson: { studentId },
    }, id, type)
  },
}))

function getRegistrationQueryModel() {
  return {
    orderNumber: registrationFilters.value.orderNumber.trim(),
    lessonId: registrationFilters.value.lessonId,
    studentId: registrationFilters.value.studentId,
    createdTimeBegin: registrationFilters.value.createdTimeBegin,
    createdTimeEnd: registrationFilters.value.createdTimeEnd,
  }
}

function getLessonQueryModel() {
  return {
    lessonId: lessonFilters.value.lessonId,
    studentId: lessonFilters.value.studentId,
  }
}

async function loadRegistrationTab() {
  const queryModel = getRegistrationQueryModel()
  const pageRequestModel = {
    needTotal: true,
    pageSize: registrationPagination.value.pageSize,
    pageIndex: registrationPagination.value.current,
    skipCount: (registrationPagination.value.current - 1) * registrationPagination.value.pageSize,
  }
  const [listRes, statsRes] = await Promise.all([
    getStudentRegistrationArrearPagedListApi({
      queryModel,
      pageRequestModel,
    }),
    getStudentRegistrationArrearStatisticsApi(queryModel),
  ])
  if (listRes.code !== 200)
    throw new Error(listRes.message || '加载报名欠费列表失败')
  if (statsRes.code !== 200)
    throw new Error(statsRes.message || '加载报名欠费统计失败')
  const resultData = listRes.result || {}
  registrationList.value = Array.isArray(resultData.list) ? resultData.list : []
  registrationPagination.value.total = Number(resultData.total || listRes.total || 0)
  registrationSummary.value.total = registrationPagination.value.total
  registrationSummary.value.totalArrearAmount = Number(statsRes.result?.totalArrearAmount || 0)
}

async function loadLessonTab() {
  const queryModel = getLessonQueryModel()
  const pageRequestModel = {
    needTotal: true,
    pageSize: lessonPagination.value.pageSize,
    pageIndex: lessonPagination.value.current,
    skipCount: (lessonPagination.value.current - 1) * lessonPagination.value.pageSize,
  }
  const [listRes, statsRes] = await Promise.all([
    getStudentLessonArrearPagedListApi({
      queryModel,
      pageRequestModel,
    }),
    getStudentLessonArrearStatisticsApi(queryModel),
  ])
  if (listRes.code !== 200)
    throw new Error(listRes.message || '加载课消欠费列表失败')
  if (statsRes.code !== 200)
    throw new Error(statsRes.message || '加载课消欠费统计失败')
  const resultData = listRes.result || {}
  lessonList.value = Array.isArray(resultData.list) ? resultData.list : []
  lessonPagination.value.total = Number(resultData.total || listRes.total || 0)
  lessonSummary.value.total = lessonPagination.value.total
  lessonSummary.value.totalArrearAmount = Number(statsRes.result?.totalArrearAmount || 0)
  lessonSummary.value.totalArrearTime = Number(statsRes.result?.totalArrearTime || 0)
}

async function getList(id?: string | number, type?: string) {
  loading.value = true
  try {
    if (activeTab.value === 'registration')
      await loadRegistrationTab()
    else
      await loadLessonTab()
    if (type)
      allFilterRef.value?.clearQuickFilter?.(id, type)
  }
  catch (error: any) {
    console.error('load arrear student list failed', error)
    messageService.error(error?.message || '加载欠费学员列表失败')
  }
  finally {
    loading.value = false
  }
}

async function initializeDefaultTab() {
  loading.value = true
  try {
    await loadRegistrationTab()
    if (registrationSummary.value.total > 0) {
      initializedDefaultTab.value = true
      return
    }

    await loadLessonTab()
    if (lessonSummary.value.total > 0)
      activeTab.value = 'lesson'

    initializedDefaultTab.value = true
  }
  catch (error: any) {
    console.error('initialize arrear tab failed', error)
    messageService.error(error?.message || '加载欠费学员列表失败')
  }
  finally {
    loading.value = false
  }
}

function handleTabChange(key: string) {
  activeTab.value = key as ArrearTabKey
  selectedRowKeys.value = []
}

function handleTableChange(pagination: { current?: number; pageSize?: number }) {
  if (activeTab.value === 'registration') {
    registrationPagination.value.current = Number(pagination.current || 1)
    registrationPagination.value.pageSize = Number(pagination.pageSize || registrationPagination.value.pageSize)
  }
  else {
    lessonPagination.value.current = Number(pagination.current || 1)
    lessonPagination.value.pageSize = Number(pagination.pageSize || lessonPagination.value.pageSize)
  }
  void getList()
}

function handleMessageRecord() {
  messageService.info('消息记录功能待接入')
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
    .filter((item): item is StudentArrearExportConditionItem => Boolean(item))

  exportModalConditionItems.value = [...mappedConditions]
  exportConditionItems.value = [...mappedConditions]
}

function getExportRecordDisplayConditions(record: StudentArrearExportRecord) {
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
  exportReportType.value = activeTab.value
  exportModalVisible.value = true
}

async function openExportRecordModal() {
  syncExportConditions()
  exportRecordModalVisible.value = true
  await loadExportRecords()
}

function parseAttachmentFilenameFromHeader(contentDisposition?: string) {
  if (!contentDisposition)
    return undefined
  const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8Match?.[1]) {
    try {
      return decodeURIComponent(utf8Match[1])
    }
    catch {
      return utf8Match[1]
    }
  }
  const plainMatch = contentDisposition.match(/filename=\"?([^\";]+)\"?/i)
  return plainMatch?.[1]
}

function triggerBlobDownload(response: any, fallbackName: string) {
  const blob = new Blob([response.data], {
    type: response.headers['content-type'] || 'application/octet-stream',
  })
  const filename = parseAttachmentFilenameFromHeader(response.headers['content-disposition']) || fallbackName
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

async function loadExportRecords() {
  exportRecordsLoading.value = true
  try {
    const res = activeTab.value === 'registration'
      ? await getStudentRegistrationArrearExportRecordsApi()
      : await getStudentLessonArrearExportRecordsApi()
    if (res.code !== 200) {
      throw new Error(res.message || '获取导出记录失败')
    }
    exportRecords.value = Array.isArray(res.result) ? res.result : []
  }
  catch (error: any) {
    console.error('load arrear export records failed', error)
    messageService.error(error?.message || '获取导出记录失败')
  }
  finally {
    exportRecordsLoading.value = false
  }
}

async function downloadExportRecord(record: StudentArrearExportRecord) {
  try {
    const response = record.exportType === 'lesson'
      ? await downloadStudentLessonArrearExportRecordApi(record.id)
      : await downloadStudentRegistrationArrearExportRecordApi(record.id)
    const fallbackName = `${record.exportType === 'lesson' ? '课消欠费' : '报名欠费'}-${dayjs().format('YYYYMMDDHHmmss')}.xlsx`
    triggerBlobDownload(response, fallbackName)
  }
  catch (error: any) {
    console.error('download arrear export record failed', error)
    messageService.error(error?.message || '下载失败，请稍后重试')
  }
}

async function handleViewExportRecord() {
  exportModalVisible.value = false
  await openExportRecordModal()
}

async function handleSubmitExport() {
  const hasData = activeTab.value === 'registration'
    ? registrationSummary.value.total > 0 && registrationList.value.length > 0
    : lessonSummary.value.total > 0 && lessonList.value.length > 0
  if (!hasData) {
    messageService.error(`没有符合条件的${currentExportName.value}数据可以导出`)
    return
  }
  exportSubmitting.value = true
  try {
    syncExportConditions()
    const res = activeTab.value === 'registration'
      ? await createStudentRegistrationArrearExportRecordApi({
          queryModel: getRegistrationQueryModel(),
          queryConditions: exportConditionItems.value,
        })
      : await createStudentLessonArrearExportRecordApi({
          queryModel: getLessonQueryModel(),
          queryConditions: exportConditionItems.value,
        })

    if (res.code !== 200) {
      throw new Error(res.message || '导出失败')
    }
    const record = res.result
    if (!record?.id) {
      throw new Error(res.message || '导出失败')
    }

    const response = activeTab.value === 'registration'
      ? await downloadStudentRegistrationArrearExportRecordApi(record.id)
      : await downloadStudentLessonArrearExportRecordApi(record.id)
    triggerBlobDownload(response, `${currentExportName.value}-${dayjs().format('YYYYMMDDHHmmss')}.xlsx`)
    exportModalVisible.value = false
    await openExportRecordModal()
  }
  catch (error: any) {
    console.error('submit arrear export failed', error)
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

function handleWechatRemind() {
  if (!selectedCount.value) {
    messageService.info('请先选择数据')
    return
  }
  messageService.info('微信提醒功能待接入')
}

function handleSmsRemind() {
  if (!selectedCount.value) {
    messageService.info('请先选择数据')
    return
  }
  messageService.info('短信提醒功能待接入')
}

async function handleBatchSetBadDebt() {
  if (activeTab.value !== 'registration')
    return
  if (!selectedCount.value) {
    messageService.info('请先选择数据')
    return
  }
  Modal.confirm({
    title: '批量设为坏账',
    content: `确认将选中的 ${selectedCount.value} 条报名欠费设为坏账吗？设为坏账后，该订单的欠费将不再追缴。`,
    okText: '确认',
    cancelText: '取消',
    async onOk() {
      if (batchOperating.value)
        return
      batchOperating.value = true
      try {
        const results = await Promise.allSettled(
          selectedRowKeys.value.map(orderId => setBadDebtApi({ orderId })),
        )
        const successCount = results.filter(item => item.status === 'fulfilled').length
        const failCount = results.length - successCount
        if (successCount > 0) {
          messageService.success(failCount > 0 ? `已设为坏账 ${successCount} 条，失败 ${failCount} 条` : `已设为坏账 ${successCount} 条`)
          selectedRowKeys.value = []
          await getList()
          return
        }
        const firstRejected = results.find(item => item.status === 'rejected')
        const errorMessage = firstRejected?.status === 'rejected' ? (firstRejected.reason?.message || '批量设为坏账失败') : '批量设为坏账失败'
        messageService.error(errorMessage)
      }
      finally {
        batchOperating.value = false
      }
    },
  })
}

function handleRegistrationAction() {
  messageService.info('补费功能待接入')
}

function handleLessonAction() {
  messageService.info('清算功能待接入')
}

watch(activeTab, async () => {
  selectedRowKeys.value = []
  exportReportType.value = activeTab.value
  if (!initializedDefaultTab.value)
    return
  if (activeTab.value === 'registration' && registrationList.value.length === 0) {
    registrationPagination.value.current = 1
    await getList()
    return
  }
  if (activeTab.value === 'lesson' && lessonList.value.length === 0) {
    lessonPagination.value.current = 1
    await getList()
    return
  }
  if (exportRecordModalVisible.value)
    await loadExportRecords()
})

useStudentListRefresh(getList)

onMounted(async () => {
  exportReportType.value = activeTab.value
  await initializeDefaultTab()
})

defineExpose({
  getList,
})
</script>

<template>
  <div class="arrear-student-page mt-2">
    <div class="arrear-panel arrear-panel--filter">
      <a-tabs :active-key="activeTab" class="arrear-tabs" @change="handleTabChange">
        <a-tab-pane key="registration" tab="报名欠费" />
        <a-tab-pane key="lesson" tab="课消欠费" />
      </a-tabs>

      <div class="filter-wrap bg-white pl-3 pr-3">
        <all-filter
          ref="allFilterRef"
          :display-array="currentDisplayArray"
          :is-quick-show="false"
          :is-show-search-stu-phonefilter="true"
          v-on="filterUpdateHandlers"
        />
      </div>
    </div>

    <div class="arrear-panel arrear-panel--table px-6 pb-5">
      <div class="table-head">
        <div class="total-text">
          {{ currentSummaryText }}
        </div>
        <div class="actions">
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
            <a-button :loading="exporting">
              导出数据
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
          <a-dropdown v-if="activeTab === 'registration'">
            <template #overlay>
              <a-menu>
                <a-menu-item key="wechat-remind" @click="handleWechatRemind">
                  微信提醒
                </a-menu-item>
                <a-menu-item key="sms-remind" @click="handleSmsRemind">
                  短信提醒
                </a-menu-item>
                <a-menu-item key="set-bad-debt" @click="handleBatchSetBadDebt">
                  设为坏账
                </a-menu-item>
              </a-menu>
            </template>
            <a-button :loading="batchOperating">
              批量操作{{ selectedCount > 0 ? `(${selectedCount})` : '' }}
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
          <a-dropdown v-else>
            <template #overlay>
              <a-menu>
                <a-menu-item key="wechat-remind" @click="handleWechatRemind">
                  微信提醒
                </a-menu-item>
                <a-menu-item key="sms-remind" @click="handleSmsRemind">
                  短信提醒
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              批量发送欠费提醒{{ selectedCount > 0 ? `(${selectedCount})` : '' }}
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
        </div>
      </div>

      <a-table
        :loading="loading"
        :columns="activeColumns"
        :data-source="activeDataSource"
        :row-key="(record: ArrearRecord) => {
          if (activeTab === 'registration')
            return 'orderId' in record ? record.orderId : ''
          return 'tuitionAccountId' in record ? `${record.studentId}-${record.lessonId}-${record.tuitionAccountId}` : record.studentId
        }"
        :row-selection="rowSelection"
        :scroll="{ x: totalWidth }"
        :pagination="{
          current: activePagination.current,
          pageSize: activePagination.pageSize,
          total: activePagination.total,
          showSizeChanger: false,
          showQuickJumper: false,
        }"
        size="small"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'student'">
            <student-avatar
              :id="record.studentId"
              :name="record.studentName || '-'"
              :gender="genderText(record.sex)"
              :avatar-url="record.avatar || undefined"
              :show-age="false"
              default-active-key="0"
            />
          </template>

          <template v-if="column.key === 'phone'">
            <div class="cell-phone">
              {{ record.phone || '-' }}
            </div>
          </template>

          <template v-if="column.key === 'arrearAmount'">
            {{ formatCurrency(record.arrearAmount) }}
          </template>

          <template v-if="column.key === 'orderAmount'">
            {{ formatCurrency(record.orderAmount) }}
          </template>

          <template v-if="column.key === 'paidAmount'">
            {{ formatCurrency(record.paidAmount) }}
          </template>

          <template v-if="column.key === 'orderNumber'">
            <clamped-text :lines="2" :text="record.orderNumber || '-'" />
          </template>

          <template v-if="column.key === 'productName'">
            <clamped-text :lines="2" :text="record.productName || '-'" />
          </template>

          <template v-if="column.key === 'createdTime'">
            {{ formatDateTime(record.createdTime) }}
          </template>

          <template v-if="column.key === 'lessonName'">
            <clamped-text :lines="2" :text="record.lessonName || '-'" />
          </template>

          <template v-if="column.key === 'arrearItem'">
            <div class="arrear-item-cell">
              <span>{{ formatLessonArrearItemParts(record).valueText }}</span>
              <span class="arrear-item-cell__sub">{{ formatLessonArrearItemParts(record).labelText }}</span>
            </div>
          </template>

          <template v-if="column.key === 'recordCount'">
            {{ record.recordCount || 0 }}条
          </template>

          <template v-if="column.key === 'action'">
            <button
              v-if="activeTab === 'registration'"
              type="button"
              class="action-link"
              @click="handleRegistrationAction"
            >
              补费
            </button>
            <button
              v-else
              type="button"
              class="action-link"
              @click="handleLessonAction"
            >
              清算
            </button>
          </template>
        </template>
      </a-table>
    </div>

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
            <a-radio :value="activeTab">
              {{ currentExportName }}
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
                :data-source="currentExportPreviewRows"
                :columns="currentExportPreviewColumns"
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
.arrear-student-page {
  overflow: hidden;
}

.arrear-panel {
  overflow: hidden;
  background: #fff;
  border-radius: 16px;
}

.arrear-panel--table {
  margin-top: 8px;
}

.arrear-tabs {
  :deep(.ant-tabs-nav) {
    margin: 0;
    padding: 0 12px;
  }
}

.table-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 18px 0 16px;
}

.total-text {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    position: absolute;
    left: 0;
    width: 4px;
    height: 12px;
    border-radius: 2px;
    background: var(--pro-ant-color-primary);
    content: "";
  }
}

.actions {
  display: flex;
  align-items: center;
}

.cell-phone {
  color: #222;
}

.arrear-item-cell {
  display: flex;
  flex-direction: column;
  line-height: 1.5;
}

.arrear-item-cell__sub {
  color: #888;
  font-size: 12px;
}

.action-link {
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--pro-ant-color-primary);
  cursor: pointer;
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
