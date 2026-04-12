<script setup lang="ts">
import dayjs from 'dayjs'
import { computed, ref, watch } from 'vue'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { getStudentTeachingRecordPagedListApi, type StudentTeachingRecordItem } from '@/api/edu-center/class-record'
import { useStudentStore } from '@/stores/student'

interface FilterOption {
  id: string
  value: string
}

const studentStore = useStudentStore()

const displayArray = ref([
  'studentIdentity',
  'classStatus',
  'billingMode',
  'scheduleCourse',
])
const studentIdentityOptions = [
  { id: 'class', value: '班级学员' },
  { id: 'one_to_one', value: '1对1学员' },
  { id: 'trial', value: '试听学员' },
  { id: 'temporary', value: '临时学员' },
  { id: 'makeup', value: '补课学员' },
]
const classStatusOptions = [
  { id: '1', value: '到课' },
  { id: '3', value: '请假' },
  { id: '2', value: '旷课' },
  { id: '0', value: '未记录' },
]
const defaultClassStatusVals: string[] = []
const lessonChargingModeOptions = [
  { id: 1, value: '按课时' },
  { id: 2, value: '按时段' },
  { id: 3, value: '按金额' },
  { id: 4, value: '不记课时' },
]
const dataSource = ref<StudentTeachingRecordItem[]>([])
const loading = ref(false)
const summary = ref({
  total: 0,
  totalClassTimes: 0,
  totalTuition: 0,
})
const courseOptions = ref<FilterOption[]>([])
const courseFinished = ref(false)
const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
})
const filterStudentIdentityValues = ref<string[]>([])
const filterClassStatusValues = ref<string[]>([...defaultClassStatusVals])
const filterLessonChargingModeValues = ref<number[]>([])
const filterLessonId = ref<string | undefined>(undefined)

const allColumns = ref<any[]>([
  {
    title: '上课日期/时段',
    dataIndex: 'classDateTime',
    key: 'classDateTime',
    fixed: 'left',
    width: 180,
  },
  {
    title: '所属班级/1v1',
    key: 'linkClass1v1',
    dataIndex: 'cloud',
    width: 180,
  },
  {
    title: '所属课程',
    key: 'course',
    dataIndex: 'course',
    width: 160,
  },
  {
    title: '科目',
    dataIndex: 'subject',
    key: 'subject',
    width: 110,
  },
  {
    title: '日程类型',
    dataIndex: 'scheduleType',
    key: 'scheduleType',
    width: 140,
  },
  {
    title: '学员身份',
    dataIndex: 'studentIdentity',
    key: 'studentIdentity',
    width: 140,
  },
  {
    title: '上课状态',
    dataIndex: 'classStatus',
    key: 'classStatus',
    width: 120,
  },
  {
    title: '扣费课程账户',
    dataIndex: 'deductionAccount',
    key: 'deductionAccount',
    width: 160,
  },
  {
    title: '课消方式',
    key: 'courseNotMethod',
    dataIndex: 'courseNotMethod',
    width: 110,
  },
  {
    title: '上课点名数量',
    dataIndex: 'classCallNum',
    key: 'classCallNum',
    width: 160,
  },
  {
    title: '消耗数量',
    dataIndex: 'useNum',
    key: 'useNum',
    width: 140,
  },
  {
    title: '拖欠数量',
    dataIndex: 'oweNum',
    key: 'oweNum',
    width: 140,
  },
  {
    title: '消耗学费',
    dataIndex: 'usePrice',
    key: 'usePrice',
    width: 140,
  },
  {
    title: '上课老师',
    dataIndex: 'mainTeacher',
    key: 'mainTeacher',
    width: 140,
  },
  {
    title: '上课助教',
    dataIndex: 'subTeacher',
    key: 'subTeacher',
    width: 140,
  },
  {
    title: '点名更新时间',
    key: 'callupdateTime',
    dataIndex: 'callupdateTime',
    width: 200,
  },
  {
    title: '对内备注',
    dataIndex: 'externalRemarks',
    key: 'externalRemarks',
    width: 140,
  },
  {
    title: '对外备注',
    dataIndex: 'remarks',
    key: 'remarks',
    width: 140,
  },
])

const totalWidth = computed(() =>
  allColumns.value.reduce((acc, column) => acc + (column.width || 0), 0),
)

const tablePagination = computed(() => ({
  current: pagination.value.current,
  pageSize: pagination.value.pageSize,
  total: pagination.value.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

function formatNumber(value?: number, suffix = '') {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return suffix ? `0${suffix}` : '0'
  const text = Number.isInteger(num) ? String(num) : num.toFixed(2).replace(/\.?0+$/, '')
  return suffix ? `${text}${suffix}` : text
}

function formatCurrency(value?: number) {
  return `¥ ${Number(value || 0).toFixed(2)}`
}

function formatDateTimeRange(record: Partial<StudentTeachingRecordItem>) {
  const start = dayjs(record.startTime)
  const end = dayjs(record.endTime)
  if (!start.isValid() || !end.isValid()) {
    return {
      dateText: '-',
      timeText: '--:-- ~ --:--',
    }
  }
  const weekday = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'][start.day()] || ''
  return {
    dateText: `${start.format('YYYY-MM-DD')} (${weekday})`,
    timeText: `${start.format('HH:mm')} ~ ${end.format('HH:mm')}`,
  }
}

function sourceTypeText(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return '临时学员'
  if (type === 3 || type === 7)
    return '补课学员'
  if (type === 4)
    return '试听学员'
  if (type === 6)
    return '1对1学员'
  return '班级学员'
}

function statusText(value?: number) {
  const status = Number(value || 0)
  if (status === 2)
    return '旷课'
  if (status === 3)
    return '请假'
  if (status === 4)
    return '未记录'
  return '到课'
}

function statusTagClass(value?: number) {
  const status = Number(value || 0)
  if (status === 2)
    return 'record-status-tag record-status-tag--absent'
  if (status === 3)
    return 'record-status-tag record-status-tag--leave'
  if (status === 4)
    return 'record-status-tag record-status-tag--pending'
  return 'record-status-tag record-status-tag--arrived'
}

function scheduleTypeText(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return '1对1日程'
  if (type === 3)
    return '试听日程'
  return '班级日程'
}

function scheduleTypeTagClass(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return 'record-meta-tag record-meta-tag--one-to-one'
  if (type === 3)
    return 'record-meta-tag record-meta-tag--trial'
  return 'record-meta-tag record-meta-tag--group'
}

function studentIdentityTagClass(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return 'record-meta-tag record-meta-tag--temporary'
  if (type === 3 || type === 7)
    return 'record-meta-tag record-meta-tag--make-up'
  if (type === 4)
    return 'record-meta-tag record-meta-tag--trial-student'
  if (type === 6)
    return 'record-meta-tag record-meta-tag--one-to-one-student'
  return 'record-meta-tag record-meta-tag--class-student'
}

function chargingModeText(value?: number) {
  const mode = Number(value || 0)
  if (mode === 2)
    return '按时间'
  if (mode === 3)
    return '按金额'
  return '按课时'
}

function isTimeChargingMode(record: Partial<StudentTeachingRecordItem>) {
  return Number(record.skuMode || 0) === 2
}

function isTrialStudent(record: Partial<StudentTeachingRecordItem>) {
  return Number(record.sourceType || 0) === 4
}

function hasArrearQuantity(record: Partial<StudentTeachingRecordItem>) {
  return Number(record.arrearQuantity || 0) > 0
}

function classDisplay(record: Partial<StudentTeachingRecordItem>) {
  return record.className || record.one2OneName || '-'
}

function mergeFilterOptions(previous: FilterOption[], incoming: FilterOption[], selectedValue?: string) {
  const selectedId = String(selectedValue || '').trim()
  const map = new Map<string, FilterOption>()
  previous.forEach((item) => {
    if (selectedId && item.id === selectedId)
      map.set(item.id, item)
  })
  incoming.forEach((item) => {
    if (item.id)
      map.set(item.id, item)
  })
  return [...map.values()]
}

function buildStudentSourceTypes(values: string[]) {
  const result = new Set<number>()
  values.forEach((item) => {
    if (item === 'class')
      result.add(5)
    else if (item === 'one_to_one')
      result.add(6)
    else if (item === 'trial')
      result.add(4)
    else if (item === 'temporary')
      result.add(2)
    else if (item === 'makeup') {
      result.add(3)
      result.add(7)
    }
  })
  return [...result]
}

function buildClassStatusValues(values: string[]) {
  return values.map((item) => {
    if (item === '0')
      return 0
    return Number(item)
  }).filter(item => Number.isFinite(item))
}

function buildQueryModel() {
  return {
    studentId: String(studentStore.studentId || '').trim() || undefined,
    lessonIds: filterLessonId.value ? [filterLessonId.value] : undefined,
    studentSourceTypes: buildStudentSourceTypes(filterStudentIdentityValues.value),
    studentTeachingRecordStatuses: filterClassStatusValues.value.length
      ? buildClassStatusValues(filterClassStatusValues.value)
      : undefined,
    lessonChargingModeEnums: filterLessonChargingModeValues.value.length ? filterLessonChargingModeValues.value : undefined,
  }
}

async function loadCourseOptions(searchKey = '', reset = true) {
  try {
    const res = await getCourseIdAndNameApi({ searchKey })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    courseOptions.value = mergeFilterOptions(
      reset ? [] : courseOptions.value,
      resultData,
      filterLessonId.value,
    )
    courseFinished.value = true
  }
  catch (error) {
    console.error('load student class record course options failed', error)
  }
}

async function loadList() {
  const studentId = String(studentStore.studentId || '').trim()
  if (!studentId) {
    dataSource.value = []
    summary.value = { total: 0, totalClassTimes: 0, totalTuition: 0 }
    pagination.value.total = 0
    return
  }
  loading.value = true
  try {
    const res = await getStudentTeachingRecordPagedListApi({
      queryModel: buildQueryModel(),
      pageRequestModel: {
        needTotal: true,
        pageIndex: pagination.value.current,
        pageSize: pagination.value.pageSize,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        startTime: 2,
        updatedTime: 0,
      },
    })
    if (res.code === 200) {
      dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
      summary.value = {
        total: Number(res.result?.total || 0),
        totalClassTimes: Number(res.result?.totalClassTimes || 0),
        totalTuition: Number(res.result?.totalTuition || 0),
      }
      pagination.value.total = Number(res.result?.total || 0)
      return
    }
    dataSource.value = []
    summary.value = { total: 0, totalClassTimes: 0, totalTuition: 0 }
    pagination.value.total = 0
  }
  catch (error) {
    console.error('load student teaching records failed', error)
    dataSource.value = []
    summary.value = { total: 0, totalClassTimes: 0, totalTuition: 0 }
    pagination.value.total = 0
  }
  finally {
    loading.value = false
  }
}

function handleStudentIdentityFilter(value: unknown) {
  filterStudentIdentityValues.value = Array.isArray(value) ? value.map(item => String(item || '')).filter(Boolean) : []
}

function handleClassStatusFilter(value: unknown) {
  filterClassStatusValues.value = Array.isArray(value) ? value.map(item => String(item || '')).filter(Boolean) : []
}

function handleLessonChargingModeFilter(value: unknown) {
  filterLessonChargingModeValues.value = Array.isArray(value)
    ? value.map(item => Number(item)).filter(item => Number.isFinite(item))
    : []
}

function handleLessonFilter(value: unknown) {
  const text = String(value || '').trim()
  filterLessonId.value = text || undefined
}

async function onScheduleCourseDropdownVisibleChange() {
  courseFinished.value = false
  await loadCourseOptions('', true)
}

async function onScheduleCourseSearch(keyword: string) {
  courseFinished.value = false
  await loadCourseOptions(keyword || '', true)
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  pagination.value.current = Number(page?.current || 1)
  pagination.value.pageSize = Number(page?.pageSize || pagination.value.pageSize)
  loadList()
}

watch(
  [
    () => String(studentStore.studentId || '').trim(),
    filterStudentIdentityValues,
    filterClassStatusValues,
    filterLessonChargingModeValues,
    filterLessonId,
  ],
  () => {
    pagination.value.current = 1
    loadList()
  },
  { deep: true, immediate: true },
)
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <all-filter
        :display-array="displayArray"
        :default-class-status-vals="defaultClassStatusVals"
        :student-identity-options="studentIdentityOptions"
        :class-status-options="classStatusOptions"
        :billing-mode-label="'课消方式'"
        :billing-mode-options-data="lessonChargingModeOptions"
        :schedule-course-options="courseOptions"
        :schedule-course-label="'所属课程'"
        :schedule-course-finished="courseFinished"
        :on-schedule-course-dropdown-visible-change="onScheduleCourseDropdownVisibleChange"
        :on-schedule-course-search="onScheduleCourseSearch"
        :is-quick-show="false"
        :student-status="1"
        @update:student-identity-filter="handleStudentIdentityFilter"
        @update:class-status-filter="handleClassStatusFilter"
        @update:billing-mode-filter="handleLessonChargingModeFilter"
        @update:charging-method-filter="handleLessonChargingModeFilter"
        @update:schedule-course-filter="handleLessonFilter"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ summary.total }} 条上课记录 学员总计 {{ formatNumber(summary.totalClassTimes, '课时') }}，共消耗学费 {{ formatCurrency(summary.totalTuition) }}
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            row-key="studentTeachingRecordId"
            :loading="loading"
            :data-source="dataSource"
            :pagination="pagination.total > pagination.pageSize ? tablePagination : false"
            :columns="allColumns"
            :scroll="{ x: totalWidth }"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'classDateTime'">
                <div class="name">
                  <div class="text-#000">
                    {{ formatDateTimeRange(record).dateText }}
                  </div>
                  <div class="text-3 text-#888 flex flex-items-center">
                    {{ formatDateTimeRange(record).timeText }}
                  </div>
                </div>
              </template>
              <template v-else-if="column.key === 'linkClass1v1'">
                {{ classDisplay(record) }}
              </template>
              <template v-else-if="column.key === 'course'">
                {{ record.lessonName || '-' }}
              </template>
              <template v-else-if="column.key === 'subject'">
                {{ record.subjectName || '-' }}
              </template>
              <template v-else-if="column.key === 'scheduleType'">
                <span :class="scheduleTypeTagClass(record.timetableSourceType)">
                  {{ scheduleTypeText(record.timetableSourceType) }}
                </span>
              </template>
              <template v-else-if="column.key === 'studentIdentity'">
                <span :class="studentIdentityTagClass(record.sourceType)">
                  {{ sourceTypeText(record.sourceType) }}
                </span>
              </template>
              <template v-else-if="column.key === 'classStatus'">
                <span :class="statusTagClass(record.status)">
                  {{ statusText(record.status) }}
                </span>
              </template>
              <template v-else-if="column.key === 'deductionAccount'">
                {{ isTrialStudent(record) ? '-' : (record.tuitionAccountName || '-') }}
              </template>
              <template v-else-if="column.key === 'courseNotMethod'">
                {{ isTrialStudent(record) ? '-' : chargingModeText(record.skuMode) }}
              </template>
              <template v-else-if="column.key === 'classCallNum'">
                {{ isTrialStudent(record) || isTimeChargingMode(record) ? '不记课时' : formatNumber(record.quantity, '课时') }}
              </template>
              <template v-else-if="column.key === 'useNum'">
                {{ isTrialStudent(record) || isTimeChargingMode(record) ? '-' : formatNumber(record.actualQuantity, '课时') }}
              </template>
              <template v-else-if="column.key === 'oweNum'">
                <span :class="{ 'owe-num-text': hasArrearQuantity(record) }">
                  {{ isTrialStudent(record) || isTimeChargingMode(record) ? '-' : formatNumber(record.arrearQuantity, '课时') }}
                </span>
              </template>
              <template v-else-if="column.key === 'usePrice'">
                <span class="use-price-text">
                  {{ isTrialStudent(record) || isTimeChargingMode(record) ? '-' : formatCurrency(record.actualTuition) }}
                </span>
              </template>
              <template v-else-if="column.key === 'mainTeacher'">
                {{ record.teacherName || '-' }}
              </template>
              <template v-else-if="column.key === 'subTeacher'">
                {{ record.assistants || '-' }}
              </template>
              <template v-else-if="column.key === 'callupdateTime'">
                <div class="update-time-cell">
                  <div>
                    {{ record.updatedTime ? dayjs(record.updatedTime).format('YYYY-MM-DD HH:mm') : '-' }}
                  </div>
                  <div class="update-time-operator">
                    {{ record.updatedStaffName || '-' }}
                  </div>
                </div>
              </template>
              <template v-else-if="column.key === 'externalRemarks'">
                {{ record.remark || '-' }}
              </template>
              <template v-else-if="column.key === 'remarks'">
                {{ record.externalRemark || '-' }}
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
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

.update-time-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.update-time-operator {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.record-status-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 48px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 20px;
  white-space: nowrap;
}

.record-meta-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 20px;
  white-space: nowrap;
  border: 1px solid transparent;
}

.record-meta-tag--group,
.record-meta-tag--class-student {
  color: #166534;
  background: #f0fdf4;
  border-color: #bbf7d0;
}

.record-meta-tag--one-to-one,
.record-meta-tag--one-to-one-student {
  color: #1d4ed8;
  background: #eff6ff;
  border-color: #bfdbfe;
}

.record-meta-tag--trial,
.record-meta-tag--trial-student {
  color: #c2410c;
  background: #fff7ed;
  border-color: #fed7aa;
}

.record-meta-tag--temporary {
  color: #7c3aed;
  background: #f5f3ff;
  border-color: #ddd6fe;
}

.record-meta-tag--make-up {
  color: #0f766e;
  background: #f0fdfa;
  border-color: #99f6e4;
}

.record-status-tag--arrived {
  color: #2f6bff;
  background: #eef4ff;
}

.record-status-tag--leave {
  color: #fa8c16;
  background: #fff4e8;
}

.record-status-tag--absent {
  color: #f5222d;
  background: #fff1f0;
}

.record-status-tag--pending {
  color: #8c8c8c;
  background: #f5f5f5;
}

.use-price-text {
  font-weight: 600;
}

.owe-num-text {
  color: #f5222d;
}
</style>
