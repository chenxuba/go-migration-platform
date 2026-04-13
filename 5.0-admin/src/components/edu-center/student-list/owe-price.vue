<script setup lang="ts">
import { DownOutlined } from '@ant-design/icons-vue'
import type { TableColumnsType } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, onMounted, ref, watch } from 'vue'
import {
  getStudentLessonArrearPagedListApi,
  getStudentLessonArrearStatisticsApi,
  getStudentRegistrationArrearPagedListApi,
  getStudentRegistrationArrearStatisticsApi,
  type StudentLessonArrearItem,
  type StudentRegistrationArrearItem,
} from '@/api/edu-center/student-list'
import { useStudentListRefresh } from '@/composables/useStudentListRefresh'
import messageService from '@/utils/messageService'

type ArrearTabKey = 'registration' | 'lesson'
type ArrearRecord = StudentRegistrationArrearItem | StudentLessonArrearItem

const activeTab = ref<ArrearTabKey>('registration')
const loading = ref(false)
const selectedRowKeys = ref<Array<string>>([])
const allFilterRef = ref<{
  clearQuickFilter?: (id?: string | number, type?: string) => void
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

const currentDisplayArray = computed(() => (activeTab.value === 'registration' ? registrationDisplayArray : lessonDisplayArray))
const activeColumns = computed(() => (activeTab.value === 'registration' ? registrationColumns : lessonColumns))
const activeDataSource = computed(() => (activeTab.value === 'registration' ? registrationList.value : lessonList.value))
const activePagination = computed(() => (activeTab.value === 'registration' ? registrationPagination.value : lessonPagination.value))
const totalWidth = computed(() => activeColumns.value.reduce((sum, item) => sum + Number(item.width || 0), 0))
const currentSummaryText = computed(() => {
  if (activeTab.value === 'registration')
    return `当前共 ${registrationSummary.value.total} 条数据，欠费金额 ${formatCurrency(registrationSummary.value.totalArrearAmount)}`
  return `当前共 ${lessonSummary.value.total} 条数据，欠费金额 ${formatCurrency(lessonSummary.value.totalArrearAmount)}，欠课时 ${formatLessonArrearTime(lessonSummary.value.totalArrearTime)}`
})

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: Array<string>) => {
    selectedRowKeys.value = keys
  },
}))

function genderText(value?: number) {
  return Number(value || 0) === 2 ? '女' : '男'
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
  getList(id, type)
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
  catch (error) {
    console.error('load arrear student list failed', error)
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
  getList()
}

function handleMessageRecord() {
  messageService.info('消息记录功能待接入')
}

function handleExportData() {
  messageService.info('导出功能待接入')
}

function handleBatchRemind() {
  messageService.info('批量提醒功能待接入')
}

function handleRegistrationAction() {
  messageService.info('补费功能待接入')
}

function handleLessonAction() {
  messageService.info('清算功能待接入')
}

watch(activeTab, async () => {
  selectedRowKeys.value = []
  if (activeTab.value === 'registration' && registrationList.value.length === 0) {
    registrationPagination.value.current = 1
    await getList()
    return
  }
  if (activeTab.value === 'lesson' && lessonList.value.length === 0) {
    lessonPagination.value.current = 1
    await getList()
  }
})

useStudentListRefresh(getList)

onMounted(async () => {
  await getList()
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
          <a-dropdown class="mr-2">
            <template #overlay>
              <a-menu>
                <a-menu-item key="export-current" @click="handleExportData">
                  导出数据
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              导出数据
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
          <a-dropdown>
            <template #overlay>
              <a-menu>
                <a-menu-item key="batch-remind" @click="handleBatchRemind">
                  批量发送欠费提醒
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              批量发送欠费提醒
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

.filter-wrap {
  border-top: 1px solid #f0f0f0;
  padding-top: 14px;
  padding-bottom: 10px;
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
</style>
