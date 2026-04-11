<script setup lang="ts">
import type { Dayjs } from 'dayjs'
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, createVNode, onMounted, ref, watch } from 'vue'
import { useTableColumns } from '@/composables/useTableColumns'
import { listClassroomsApi } from '@/api/business-settings/classroom'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { getRollCallPagedListApi, getRollCallStatisticsApi } from '@/api/edu-center/roll-call'
import { cancelTeachingSchedulesApi, type TeachingScheduleItem } from '@/api/edu-center/teaching-schedule'
import messageService from '@/utils/messageService'

interface FilterOption {
  id: string
  value: string
}

interface AllFilterExpose {
  setScheduleDateFilter: (values?: string[], shouldEmit?: boolean) => void
}

interface TeacherRoleFilterPayload {
  teacherType?: number
  teacherId?: string
}

type DashboardFilter = 'today' | 'all' | 'partial' | 'custom'

const today = dayjs().format('YYYY-MM-DD')
const monthStart = dayjs().startOf('month')
const todayDayjs = dayjs()
const defaultScheduleDateVals = [monthStart.format('YYYY-MM-DD'), todayDayjs.format('YYYY-MM-DD')]

const displayArray = ref([
  'scheduleDate',
  'scheduleCourse',
  'scheduleClassroom',
  'scheduleClass',
  'scheduleOneToOne',
  'scheduleType',
])

const dataSource = ref<TeachingScheduleItem[]>([])
const tableLoading = ref(false)
const statisticsLoading = ref(false)
const batchDeleting = ref(false)
const openDrawer = ref(false)
const allFilterRef = ref<AllFilterExpose | null>(null)
const dashboardFilter = ref<DashboardFilter>('custom')
const dateRange = ref<[Dayjs, Dayjs]>([monthStart, todayDayjs])
const sortDirection = ref(2)

const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
})
const selectedRowKeys = ref<string[]>([])

const dashboardStats = ref({
  todayCount: 0,
  allCount: 0,
  partialCount: 0,
})

const filterTeacherRole = ref<TeacherRoleFilterPayload>({
  teacherType: 1,
  teacherId: undefined,
})
const filterClassroomId = ref<string[]>([])
const filterClassId = ref<string | undefined>(undefined)
const filterOneToOneId = ref<string | undefined>(undefined)
const filterCourseId = ref<string | undefined>(undefined)
const filterScheduleType = ref<string[]>([])

const scheduleClassroomOptions = ref<FilterOption[]>([])
const scheduleClassOptions = ref<FilterOption[]>([])
const scheduleOneToOneOptions = ref<FilterOption[]>([])
const scheduleCourseOptions = ref<FilterOption[]>([])

const scheduleClassroomFinished = ref(false)
const scheduleClassFinished = ref(false)
const scheduleOneToOneFinished = ref(false)
const scheduleCourseFinished = ref(false)

const scheduleClassPagination = ref({ current: 1, pageSize: 20, total: 0 })
const scheduleOneToOnePagination = ref({ current: 1, pageSize: 20, total: 0 })

const scheduleClassSearchKey = ref('')
const scheduleOneToOneSearchKey = ref('')

const scheduleTypeOptions = [
  { id: 'group_class', value: '班级日程' },
  { id: 'one_to_one', value: '1对1日程' },
  { id: 'trial', value: '试听日程' },
]

const allColumns = ref([
  {
    title: '上课日期/时段',
    dataIndex: 'classDateTime',
    key: 'classDateTime',
    width: 180,
    fixed: 'left',
    sorter: true,
    defaultSortOrder: 'descend',
  },
  {
    title: '日程类型',
    dataIndex: 'scheduleType',
    key: 'scheduleType',
    width: 120,
  },
  {
    title: '班级/1对1',
    key: 'classOr1v1',
    dataIndex: 'classOr1v1',
    width: 180,
  },
  {
    title: '课程名称',
    key: 'courseName',
    dataIndex: 'courseName',
    width: 140,
  },
  {
    title: '上课老师',
    key: 'mainTeacher',
    dataIndex: 'mainTeacher',
    width: 110,
  },
  {
    title: '上课助教',
    dataIndex: 'subTeacher',
    key: 'subTeacher',
    width: 140,
  },
  {
    title: '上课教室',
    dataIndex: 'classRoom',
    key: 'classRoom',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 100,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'roll-call-list',
    allColumns,
    excludeKeys: ['action'],
  })

const tablePagination = computed(() => ({
  current: pagination.value.current,
  pageSize: pagination.value.pageSize,
  total: pagination.value.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: (string | number)[]) => {
    selectedRowKeys.value = keys.map(key => String(key))
  },
}))

function normalizeScheduleFilterValue(value: unknown) {
  if (Array.isArray(value))
    return value.length ? String(value[0] ?? '').trim() || undefined : undefined
  const text = String(value ?? '').trim()
  return text || undefined
}

function normalizeScheduleFilterValues(value: unknown) {
  if (!Array.isArray(value))
    return []
  return value.map(item => String(item ?? '').trim()).filter(Boolean)
}

function mergeFilterOptions(previous: FilterOption[], incoming: FilterOption[], selectedValues: string | string[] | undefined = []) {
  const selectedSet = new Set((Array.isArray(selectedValues) ? selectedValues : [selectedValues]).map(value => String(value || '')).filter(Boolean))
  const map = new Map<string, FilterOption>()
  previous.forEach((item) => {
    if (selectedSet.has(item.id))
      map.set(item.id, item)
  })
  incoming.forEach((item) => {
    if (item.id)
      map.set(item.id, item)
  })
  return [...map.values()]
}

function isSameDateRange(nextStart: Dayjs, nextEnd: Dayjs) {
  return dateRange.value[0]?.format('YYYY-MM-DD') === nextStart.format('YYYY-MM-DD')
    && dateRange.value[1]?.format('YYYY-MM-DD') === nextEnd.format('YYYY-MM-DD')
}

function buildStatisticsQueryModel() {
  return {
    lessonId: filterCourseId.value,
    classroomId: filterClassroomId.value[0],
    classId: filterClassId.value,
    oneToOneId: filterOneToOneId.value,
    teacherId: filterTeacherRole.value.teacherId,
    teacherTypes: filterTeacherRole.value.teacherId ? [Number(filterTeacherRole.value.teacherType || 1)] : undefined,
    scheduleTypes: filterScheduleType.value.length ? filterScheduleType.value : undefined,
  }
}

function buildDateRangeForList() {
  if (dashboardFilter.value === 'today') {
    return {
      startDate: today,
      endDate: today,
    }
  }
  if (dashboardFilter.value === 'all' || dashboardFilter.value === 'partial') {
    return {
      startDate: undefined,
      endDate: today,
    }
  }
  return {
    startDate: dateRange.value[0]?.format('YYYY-MM-DD'),
    endDate: dateRange.value[1]?.format('YYYY-MM-DD'),
  }
}

function buildListQueryModel() {
  const range = buildDateRangeForList()
  return {
    ...buildStatisticsQueryModel(),
    startDate: range.startDate,
    endDate: range.endDate,
  }
}

async function loadStatistics() {
  statisticsLoading.value = true
  try {
    const res = await getRollCallStatisticsApi({
      queryModel: buildStatisticsQueryModel(),
    })
    if (res.code === 200) {
      dashboardStats.value = {
        todayCount: Number(res.result?.todayCount || 0),
        allCount: Number(res.result?.allCount || 0),
        partialCount: Number(res.result?.partialCount || 0),
      }
      return
    }
    dashboardStats.value = { todayCount: 0, allCount: 0, partialCount: 0 }
  }
  catch (error) {
    console.error('load roll call statistics failed', error)
    dashboardStats.value = { todayCount: 0, allCount: 0, partialCount: 0 }
  }
  finally {
    statisticsLoading.value = false
  }
}

async function loadList() {
  if (dashboardFilter.value === 'partial') {
    dataSource.value = []
    pagination.value.total = 0
    selectedRowKeys.value = []
    return
  }

  tableLoading.value = true
  try {
    const res = await getRollCallPagedListApi({
      queryModel: buildListQueryModel(),
      pageRequestModel: {
        needTotal: true,
        pageIndex: pagination.value.current,
        pageSize: pagination.value.pageSize,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        byStartDate: sortDirection.value,
      },
    })
    if (res.code === 200) {
      dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
      pagination.value.total = Number(res.result?.total || 0)
      selectedRowKeys.value = selectedRowKeys.value.filter(key => dataSource.value.some(item => String(item.id) === key))
      return
    }
    dataSource.value = []
    pagination.value.total = 0
    selectedRowKeys.value = []
  }
  catch (error) {
    console.error('load roll call list failed', error)
    dataSource.value = []
    pagination.value.total = 0
    selectedRowKeys.value = []
  }
  finally {
    tableLoading.value = false
  }
}

function resetToFirstPage() {
  pagination.value.current = 1
}

function handleScheduleClassroomFilter(value: unknown) {
  filterClassroomId.value = normalizeScheduleFilterValues(value)
}

function handleScheduleClassFilter(value: unknown) {
  filterClassId.value = normalizeScheduleFilterValue(value)
}

function handleScheduleOneToOneFilter(value: unknown) {
  filterOneToOneId.value = normalizeScheduleFilterValue(value)
}

function handleScheduleCourseFilter(value: unknown) {
  filterCourseId.value = normalizeScheduleFilterValue(value)
}

function handleScheduleDateFilter(value: unknown) {
  const normalized = normalizeScheduleFilterValues(value)
  if (normalized.length >= 2) {
    const start = dayjs(normalized[0])
    const end = dayjs(normalized[1])
    if (start.isValid() && end.isValid()) {
      if (dashboardFilter.value === 'custom' && isSameDateRange(start, end))
        return
      dateRange.value = [start, end]
      dashboardFilter.value = 'custom'
      return
    }
  }
  if (dashboardFilter.value === 'custom' && isSameDateRange(monthStart, todayDayjs))
    return
  dateRange.value = [monthStart, todayDayjs]
  dashboardFilter.value = 'custom'
}

function handleScheduleTypeFilter(value: unknown) {
  filterScheduleType.value = normalizeScheduleFilterValues(value)
}

function handleTeacherRoleFilter(payload: TeacherRoleFilterPayload) {
  filterTeacherRole.value = {
    teacherType: Number(payload?.teacherType || 1),
    teacherId: payload?.teacherId ? String(payload.teacherId) : undefined,
  }
}

function handleQuickFilter(type: DashboardFilter) {
  dashboardFilter.value = type
  if (type === 'today') {
    dateRange.value = [todayDayjs, todayDayjs]
    allFilterRef.value?.setScheduleDateFilter([today, today])
  }
  else if (type === 'custom') {
    allFilterRef.value?.setScheduleDateFilter([
      dateRange.value[0].format('YYYY-MM-DD'),
      dateRange.value[1].format('YYYY-MM-DD'),
    ])
  }
}

function handleRollCall() {
  openDrawer.value = true
}

function handleBatchDelete() {
  if (!selectedRowKeys.value.length) {
    messageService.warning('请先选择要删除的日程')
    return
  }

  Modal.confirm({
    title: '确定删除？',
    icon: createVNode(ExclamationCircleOutlined, { style: 'color: #ff4d4f' }),
    content: `删除的日程无法恢复，并不会再显示在课表中。已选择 ${selectedRowKeys.value.length} 条，确认删除？`,
    okText: '确定删除',
    cancelText: '再想想',
    async onOk() {
      batchDeleting.value = true
      try {
        const res = await cancelTeachingSchedulesApi({
          ids: selectedRowKeys.value,
        })
        if (res.code !== 200)
          throw new Error(res.message || '批量删除失败')
        messageService.success(`已删除 ${Number(res.result?.canceled || selectedRowKeys.value.length)} 条日程`)
        selectedRowKeys.value = []
        await loadStatistics()
        await loadList()
      }
      catch (error: any) {
        console.error('batch delete roll call schedules failed', error)
        messageService.error(error?.response?.data?.message || error?.message || '批量删除失败')
        throw error
      }
      finally {
        batchDeleting.value = false
      }
    },
  })
}

function handleTableChange(page: { current?: number, pageSize?: number }, _filters: unknown, sorter: { order?: string } | Array<{ order?: string }>) {
  pagination.value.current = Number(page?.current || 1)
  pagination.value.pageSize = Number(page?.pageSize || pagination.value.pageSize)
  const currentSorter = Array.isArray(sorter) ? sorter[0] : sorter
  if (currentSorter?.order === 'ascend')
    sortDirection.value = 1
  else if (currentSorter?.order === 'descend')
    sortDirection.value = 2
  else
    sortDirection.value = 2
  loadList()
}

function scheduleTypeLabel(record: Record<string, any>) {
  return Number(record.classType) === 2 ? '1对1日程' : '班级日程'
}

function classDisplayName(record: Record<string, any>) {
  return record.teachingClassName || record.studentName || '-'
}

function assistantTeacherText(record: Record<string, any>) {
  if (Array.isArray(record.assistantNames) && record.assistantNames.length)
    return record.assistantNames.join('、')
  return '-'
}

function formatLessonWeekday(lessonDate?: string) {
  if (!lessonDate)
    return ''
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return weekdays[dayjs(lessonDate).day()] || ''
}

function formatLessonTime(record: Record<string, any>) {
  const dateText = record.lessonDate ? `${record.lessonDate} (${formatLessonWeekday(record.lessonDate)})` : '-'
  const startText = record.startAt ? dayjs(record.startAt).format('HH:mm') : '--:--'
  const endText = record.endAt ? dayjs(record.endAt).format('HH:mm') : '--:--'
  return {
    dateText,
    timeText: `${startText} ~ ${endText}`,
  }
}

async function loadScheduleClassroomOptions(searchKey = '') {
  try {
    const res = await listClassroomsApi({
      enabledOnly: true,
      searchKey,
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleClassroomOptions.value = mergeFilterOptions(scheduleClassroomOptions.value, resultData, filterClassroomId.value)
    scheduleClassroomFinished.value = true
  }
  catch (error) {
    console.error('load roll call classrooms failed', error)
  }
}

async function loadScheduleClassOptions(searchKey = '', reset = true) {
  if (reset) {
    scheduleClassPagination.value.current = 1
    scheduleClassFinished.value = false
  }
  scheduleClassSearchKey.value = searchKey
  try {
    const res = await pageGroupClassesApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: scheduleClassPagination.value.pageSize,
        pageIndex: scheduleClassPagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        className: searchKey || undefined,
      },
    })
    if (res.code !== 200)
      return
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    const resultData = list.map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleClassOptions.value = reset
      ? mergeFilterOptions(scheduleClassOptions.value, resultData, filterClassId.value)
      : mergeFilterOptions(scheduleClassOptions.value, [...scheduleClassOptions.value, ...resultData], filterClassId.value)
    scheduleClassPagination.value.total = Number(res.result?.total || resultData.length || 0)
    scheduleClassFinished.value = scheduleClassOptions.value.length >= scheduleClassPagination.value.total
  }
  catch (error) {
    console.error('load roll call classes failed', error)
  }
}

async function loadScheduleOneToOneOptions(searchKey = '', reset = true) {
  if (reset) {
    scheduleOneToOnePagination.value.current = 1
    scheduleOneToOneFinished.value = false
  }
  scheduleOneToOneSearchKey.value = searchKey
  try {
    const res = await getOneToOneListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: scheduleOneToOnePagination.value.pageSize,
        pageIndex: scheduleOneToOnePagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        status: [1],
        searchKey,
      },
    })
    if (res.code !== 200)
      return
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    const resultData = list.map(item => ({
      id: String(item.id ?? ''),
      value: `${String(item.studentName || item.name || item.id || '').trim()}～${String(item.lessonName || '').trim()}`.replace(/～$/, ''),
    })).filter(item => item.id && item.value)
    scheduleOneToOneOptions.value = reset
      ? mergeFilterOptions(scheduleOneToOneOptions.value, resultData, filterOneToOneId.value)
      : mergeFilterOptions(scheduleOneToOneOptions.value, [...scheduleOneToOneOptions.value, ...resultData], filterOneToOneId.value)
    scheduleOneToOnePagination.value.total = Number(res.result?.total || resultData.length || 0)
    scheduleOneToOneFinished.value = scheduleOneToOneOptions.value.length >= scheduleOneToOnePagination.value.total
  }
  catch (error) {
    console.error('load roll call one-to-one options failed', error)
  }
}

async function loadScheduleCourseOptions(searchKey = '', reset = true) {
  try {
    const res = await getCourseIdAndNameApi({ searchKey })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleCourseOptions.value = mergeFilterOptions(reset ? [] : scheduleCourseOptions.value, resultData, filterCourseId.value)
    scheduleCourseFinished.value = true
  }
  catch (error) {
    console.error('load roll call courses failed', error)
  }
}

async function onScheduleClassroomDropdownVisibleChange() {
  await loadScheduleClassroomOptions('')
}

async function onScheduleClassroomSearch(keyword: string) {
  await loadScheduleClassroomOptions(keyword || '')
}

async function onScheduleClassDropdownVisibleChange() {
  await loadScheduleClassOptions('', true)
}

async function onScheduleClassSearch(keyword: string) {
  await loadScheduleClassOptions(keyword || '', true)
}

async function loadMoreScheduleClass() {
  if (scheduleClassFinished.value)
    return
  scheduleClassPagination.value.current += 1
  await loadScheduleClassOptions(scheduleClassSearchKey.value, false)
}

async function onScheduleOneToOneDropdownVisibleChange() {
  await loadScheduleOneToOneOptions('', true)
}

async function onScheduleOneToOneSearch(keyword: string) {
  await loadScheduleOneToOneOptions(keyword || '', true)
}

async function loadMoreScheduleOneToOne() {
  if (scheduleOneToOneFinished.value)
    return
  scheduleOneToOnePagination.value.current += 1
  await loadScheduleOneToOneOptions(scheduleOneToOneSearchKey.value, false)
}

async function onScheduleCourseDropdownVisibleChange() {
  await loadScheduleCourseOptions('', true)
}

async function onScheduleCourseSearch(keyword: string) {
  await loadScheduleCourseOptions(keyword || '', true)
}

watch(
  [filterTeacherRole, filterClassroomId, filterClassId, filterOneToOneId, filterCourseId, filterScheduleType],
  () => {
    resetToFirstPage()
    loadStatistics()
    loadList()
  },
  { deep: true },
)

watch(
  [dashboardFilter, dateRange],
  () => {
    resetToFirstPage()
    loadList()
  },
  { deep: true },
)

onMounted(async () => {
  await loadStatistics()
  await loadList()
})
</script>

<template>
  <div class="roll-call">
    <div class="databord bg-white pt-3 pb-3 pl-5 pr-5 rounded-4">
      <custom-title title="关键数据看板" font-size="14px" font-weight="500" />
      <div class="flex justify-between mt-3 mb-2">
        <div class="flex-1 bg-#fbfcff h-22.5 cursor-pointer rounded-5 hover-bg-#0066ff0d" @click="handleQuickFilter('today')">
          <div class="contentMain">
            <div class="contentMainLeft">
              今日待点名
            </div>
            <div class="contentMainRight">
              {{ statisticsLoading ? '-' : dashboardStats.todayCount }}
            </div>
          </div>
          <div class="contentSub">
            <div class="contentSubLeft">
              今日待点名的日程
            </div>
            <div class="contentSubRight">
              快捷筛选
            </div>
          </div>
        </div>
        <div class="flex-1 bg-#fbfcff h-22.5 cursor-pointer rounded-5 hover-bg-#0066ff0d ml-4 mr-4" @click="handleQuickFilter('all')">
          <div class="contentMain">
            <div class="contentMainLeft">
              全部待点名
            </div>
            <div class="contentMainRight">
              {{ statisticsLoading ? '-' : dashboardStats.allCount }}
            </div>
          </div>
          <div class="contentSub">
            <div class="contentSubLeft">
              过去至今从未点名的日程
            </div>
            <div class="contentSubRight">
              快捷筛选
            </div>
          </div>
        </div>
        <div class="flex-1 bg-#fbfcff h-22.5 cursor-pointer rounded-5 hover-bg-#0066ff0d" @click="handleQuickFilter('partial')">
          <div class="contentMain">
            <div class="contentMainLeft">
              部分点名
            </div>
            <div class="contentMainRight">
              {{ statisticsLoading ? '-' : dashboardStats.partialCount }}
            </div>
          </div>
          <div class="contentSub">
            <div class="contentSubLeft">
              已点名但未完成全部点名的日程
            </div>
            <div class="contentSubRight">
              前往处理
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="bg-white rounded-4 mt-3 pl-2 pr-2 py-3">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :default-schedule-date-vals="defaultScheduleDateVals"
        :schedule-date-disable-future="true"
        is-show-chang-teacher-search
        :schedule-classroom-options="scheduleClassroomOptions"
        :schedule-classroom-finished="scheduleClassroomFinished"
        :on-schedule-classroom-dropdown-visible-change="onScheduleClassroomDropdownVisibleChange"
        :on-schedule-classroom-search="onScheduleClassroomSearch"
        :schedule-class-options="scheduleClassOptions"
        :schedule-class-finished="scheduleClassFinished"
        :on-schedule-class-dropdown-visible-change="onScheduleClassDropdownVisibleChange"
        :on-schedule-class-search="onScheduleClassSearch"
        :on-schedule-class-load-more="loadMoreScheduleClass"
        :schedule-one-to-one-options="scheduleOneToOneOptions"
        :schedule-one-to-one-finished="scheduleOneToOneFinished"
        :on-schedule-one-to-one-dropdown-visible-change="onScheduleOneToOneDropdownVisibleChange"
        :on-schedule-one-to-one-search="onScheduleOneToOneSearch"
        :on-schedule-one-to-one-load-more="loadMoreScheduleOneToOne"
        :schedule-course-options="scheduleCourseOptions"
        :schedule-course-finished="scheduleCourseFinished"
        :on-schedule-course-dropdown-visible-change="onScheduleCourseDropdownVisibleChange"
        :on-schedule-course-search="onScheduleCourseSearch"
        :schedule-type-options="scheduleTypeOptions"
        @update:schedule-classroom-filter="handleScheduleClassroomFilter"
        @update:schedule-class-filter="handleScheduleClassFilter"
        @update:schedule-one-to-one-filter="handleScheduleOneToOneFilter"
        @update:schedule-course-filter="handleScheduleCourseFilter"
        @update:schedule-date-filter="handleScheduleDateFilter"
        @update:schedule-type-filter="handleScheduleTypeFilter"
        @update:teacher-role-filter="handleTeacherRoleFilter"
      />
    </div>

    <div class="bg-white rounded-4 mt-3 py-3 px-5">
      <div class="table-title flex justify-between">
        <div class="total">
          当前共计 {{ pagination.total }} 条待点名日程
        </div>
        <div class="edit flex">
          <a-button class="mr-3" :loading="batchDeleting" @click="handleBatchDelete">
            批量删除
          </a-button>
          <a-button class="mr-3">
            批量点名
          </a-button>
          <a-button class="mr-3" type="primary">
            创建未排课点名
          </a-button>
          <customize-code
            v-model:checked-values="selectedValues"
            :options="columnOptions"
            :total="allColumns.length - 1"
            :num="selectedValues.length - 1"
          />
        </div>
      </div>
      <a-table
        row-key="id"
        :loading="tableLoading"
        :data-source="dataSource"
        :pagination="tablePagination"
        :row-selection="rowSelection"
        :columns="filteredColumns"
        :scroll="{ x: totalWidth }"
        size="small"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'classDateTime'">
            <div>{{ formatLessonTime(record).dateText }}</div>
            <div>{{ formatLessonTime(record).timeText }}</div>
          </template>
          <template v-else-if="column.key === 'scheduleType'">
            <span>{{ scheduleTypeLabel(record) }}</span>
          </template>
          <template v-else-if="column.key === 'classOr1v1'">
            {{ classDisplayName(record) }}
          </template>
          <template v-else-if="column.key === 'courseName'">
            {{ record.lessonName || '-' }}
          </template>
          <template v-else-if="column.key === 'mainTeacher'">
            {{ record.teacherName || '-' }}
          </template>
          <template v-else-if="column.key === 'subTeacher'">
            {{ assistantTeacherText(record) }}
          </template>
          <template v-else-if="column.key === 'classRoom'">
            {{ record.classroomName || '-' }}
          </template>
          <template v-else-if="column.key === 'action'">
            <span class="flex action">
              <a class="font500" @click="handleRollCall()">点名</a>
            </span>
          </template>
        </template>
      </a-table>
    </div>

    <roll-call-drawer v-model:open="openDrawer" />
  </div>
</template>

<style lang="less" scoped>
.contentMain {
  box-sizing: content-box;
  padding: 16px 12px 6px 24px;
  height: 30px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;

  .contentMainLeft {
    font-size: 14px;
    font-weight: 500;
    color: #222;
    flex-shrink: 0;
  }

  .contentMainRight {
    min-width: 72px;
    height: 30px;
    font-size: 30px;
    font-weight: 700;
    font-family: DINAlternate-Bold, DINAlternate;
    color: #06f;
    line-height: 30px;
    flex-shrink: 0;
    text-align: center;
  }
}

.contentSub {
  padding: 0 24px;
  height: 16px;
  line-height: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;

  .contentSubLeft {
    font-size: 13px;
    color: #888;
  }

  .contentSubRight {
    font-size: 12px;
    color: #06f;
  }
}

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
</style>
