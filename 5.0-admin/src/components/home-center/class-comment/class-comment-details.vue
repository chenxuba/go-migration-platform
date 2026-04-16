<script setup lang="ts">
import dayjs, { type Dayjs } from 'dayjs'
import { computed, onMounted, ref, watch } from 'vue'
import scheduleClassImage from '@/assets/images/timetable/schedule-class.png'
import scheduleOneToOneImage from '@/assets/images/timetable/schedule-one2one.png'
import {
  getClassCommentStudentPagedListApi,
  type ClassCommentStudentItem,
} from '@/api/edu-center/class-record'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import StudentAvatar from '@/components/common/StudentAvatar.vue'
import { useTableColumns } from '@/composables/useTableColumns'
import messageService from '@/utils/messageService'

interface FilterOption {
  id: string
  value: string
  mobile?: string
}

const monthStart = dayjs().startOf('month')
const today = dayjs()
const defaultScheduleDateVals = [monthStart.format('YYYY-MM-DD'), today.format('YYYY-MM-DD')]

const displayArray = ref([
  'scheduleCourse',
  'scheduleTeacher',
  'assistantTeacher',
  'scheduleClass',
  'scheduleOneToOne',
  'scheduleDate',
  'scheduleType',
  'commentStatus',
  'readStatus',
  'parentFeedbackStatus',
  'stuPhoneSearch',
])

const scheduleTypeOptions = [
  { id: '1', value: '班级' },
  { id: '2', value: '1对1' },
]

const commentStatusOptions = [
  { id: '1', value: '已点评' },
  { id: '0', value: '未点评' },
]

const readStatusOptions = [
  { id: '1', value: '已读' },
  { id: '0', value: '未读' },
]

const parentFeedbackStatusOptions = [
  { id: '1', value: '已反馈' },
  { id: '0', value: '未反馈' },
]

const loading = ref(false)
const dataSource = ref<ClassCommentStudentItem[]>([])
const sortStartTime = ref(2)

const filterDateRange = ref<[Dayjs, Dayjs]>([monthStart, today])
const filterLessonId = ref<string | undefined>(undefined)
const filterTeacherIds = ref<string[]>([])
const filterAssistantTeacherIds = ref<string[]>([])
const filterClassId = ref<string | undefined>(undefined)
const filterOneToOneId = ref<string | undefined>(undefined)
const filterScheduleTypes = ref<string[]>([])
const filterCommentStatus = ref<boolean | undefined>(undefined)
const filterReadStatus = ref<boolean | undefined>(undefined)
const filterParentFeedbackStatus = ref<boolean | undefined>(undefined)
const filterStudentId = ref<string | undefined>(undefined)

const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
})

const courseOptions = ref<FilterOption[]>([])
const courseFinished = ref(false)
const classOptions = ref<FilterOption[]>([])
const classFinished = ref(false)
const oneToOneOptions = ref<FilterOption[]>([])
const oneToOneFinished = ref(false)
const teacherOptions = ref<FilterOption[]>([])
const teacherFinished = ref(false)
const assistantTeacherOptions = ref<FilterOption[]>([])
const assistantTeacherFinished = ref(false)

const classPagination = ref({ current: 1, pageSize: 20, total: 0 })
const oneToOnePagination = ref({ current: 1, pageSize: 20, total: 0 })
const teacherPagination = ref({ current: 1, pageSize: 20, total: 0 })
const assistantTeacherPagination = ref({ current: 1, pageSize: 20, total: 0 })

const classSearchKey = ref('')
const oneToOneSearchKey = ref('')
const teacherSearchKey = ref('')
const assistantTeacherSearchKey = ref('')

const allColumns = ref([
  {
    title: '上课日期/时段',
    dataIndex: 'classDateTime',
    key: 'classDateTime',
    fixed: 'left',
    width: 160,
    sorter: true,
    defaultSortOrder: 'descend',
    required: true,
  },
  {
    title: '类型',
    dataIndex: 'type',
    key: 'type',
    width: 120,
  },
  {
    title: '所属班级/1对1',
    key: 'linkClassOr1v1',
    dataIndex: 'linkClassOr1v1',
    width: 200,
  },
  {
    title: '所属课程',
    key: 'linkCourse',
    dataIndex: 'linkCourse',
    width: 160,
  },
  {
    title: '上课教师',
    key: 'teacher',
    dataIndex: 'teacher',
    width: 130,
  },
  {
    title: '学员/电话',
    dataIndex: 'student',
    key: 'student',
    width: 170,
  },
  {
    title: '是否点评',
    key: 'isComment',
    dataIndex: 'isComment',
    width: 100,
  },
  {
    title: '已读/未读',
    dataIndex: 'readStatus',
    key: 'readStatus',
    width: 120,
  },
  {
    title: '课评反馈',
    dataIndex: 'parentFeedback',
    key: 'parentFeedback',
    width: 120,
  },
  {
    title: '反馈分数',
    dataIndex: 'parentFeedbackGrade',
    key: 'parentFeedbackGrade',
    width: 120,
  },
  {
    title: '反馈评论',
    dataIndex: 'parentFeedbackContent',
    key: 'parentFeedbackContent',
    width: 120,
  },
  {
    title: '上课助教',
    dataIndex: 'subTeacher',
    key: 'subTeacher',
    width: 150,
  },
  {
    title: '上课教室',
    dataIndex: 'classRoom',
    key: 'classRoom',
    width: 130,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 140,
  },
])

const { filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'class-comment-details',
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

function normalizeFilterValue(value: unknown) {
  if (Array.isArray(value))
    return value.length ? String(value[0] ?? '').trim() || undefined : undefined
  const text = String(value ?? '').trim()
  return text || undefined
}

function normalizeFilterValues(value: unknown) {
  if (!Array.isArray(value))
    return []
  return value.map(item => String(item ?? '').trim()).filter(Boolean)
}

function normalizeBooleanFilter(value: unknown) {
  const normalized = normalizeFilterValue(value)
  if (normalized === '1')
    return true
  if (normalized === '0')
    return false
  return undefined
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

function formatDateTime(record: Partial<ClassCommentStudentItem>) {
  const start = dayjs(record.startTime)
  const end = dayjs(record.endTime)
  if (!start.isValid() || !end.isValid()) {
    return {
      dateText: '-',
      timeText: '--:--～--:--',
    }
  }
  const weekday = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'][start.day()] || ''
  return {
    dateText: `${start.format('YYYY-MM-DD')}（${weekday}）`,
    timeText: `${start.format('HH:mm')}～${end.format('HH:mm')}`,
  }
}

function sourceTypeText(record: Partial<ClassCommentStudentItem>) {
  const sourceType = Number(record.sourceType || 0)
  if (sourceType === 2)
    return '1对1'
  if (sourceType === 3)
    return '试听'
  return '班级'
}

function sourceTypeImage(record: Partial<ClassCommentStudentItem>) {
  const sourceType = Number(record.sourceType || 0)
  if (sourceType === 2)
    return scheduleOneToOneImage
  if (sourceType === 1)
    return scheduleClassImage
  return ''
}

function displayText(value?: string | null) {
  const text = String(value ?? '').trim()
  return text || '-'
}

function readStatusText(record: Partial<ClassCommentStudentItem>) {
  if (record.isRead === true)
    return '已读'
  if (record.isRead === false)
    return '未读'
  return '-'
}

function parentFeedbackText(record: Partial<ClassCommentStudentItem>) {
  const hasFeedbackData = record.isParentFeedback === true
    || Number(record.parentFeedbackType || 0) > 0
    || Number(record.parentFeedbackGrade || 0) > 0
    || String(record.parentFeedbackContent || '').trim() !== ''
  if (!hasFeedbackData)
    return '-'
  return record.isParentFeedback ? '已反馈' : '-'
}

function parentFeedbackGradeText(record: Partial<ClassCommentStudentItem>) {
  const grade = Number(record.parentFeedbackGrade || 0)
  return grade > 0 ? `${grade}分` : '-'
}

function handleViewPending() {
  messageService.info('暂未开发')
}

function buildQueryModel() {
  return {
    isParentFeedback: filterParentFeedbackStatus.value,
    teachingStartTime: filterDateRange.value?.[0]?.format('YYYY-MM-DD'),
    teachingEndTime: filterDateRange.value?.[1]?.format('YYYY-MM-DD'),
    teachingRecordTypes: filterScheduleTypes.value.length
      ? filterScheduleTypes.value.map(item => Number(item)).filter(Boolean)
      : [1, 2],
    lessonId: filterLessonId.value,
    teacherIds: filterTeacherIds.value.length ? filterTeacherIds.value : undefined,
    assistantTeacherIds: filterAssistantTeacherIds.value.length ? filterAssistantTeacherIds.value : undefined,
    classId: filterClassId.value,
    one2OneId: filterOneToOneId.value,
    isComment: filterCommentStatus.value,
    isRead: filterReadStatus.value,
    classTeacherIds: [],
    one2OneTeacherIds: [],
    studentId: filterStudentId.value,
  }
}

async function loadCourseOptions(searchKey = '') {
  try {
    const res = await getCourseIdAndNameApi({
      searchKey: searchKey || '',
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.value || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    courseOptions.value = mergeFilterOptions(courseOptions.value, resultData, filterLessonId.value)
    courseFinished.value = true
  }
  catch (error) {
    console.error('load class comment detail courses failed', error)
  }
}

async function loadClassOptions(searchKey = '', reset = true) {
  if (reset) {
    classPagination.value.current = 1
    classFinished.value = false
  }
  classSearchKey.value = searchKey
  try {
    const res = await pageGroupClassesApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: classPagination.value.pageSize,
        pageIndex: classPagination.value.current,
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
    classOptions.value = reset
      ? mergeFilterOptions(classOptions.value, resultData, filterClassId.value)
      : mergeFilterOptions(classOptions.value, [...classOptions.value, ...resultData], filterClassId.value)
    classPagination.value.total = Number(res.result?.total || resultData.length || 0)
    classFinished.value = classOptions.value.length >= classPagination.value.total
  }
  catch (error) {
    console.error('load class comment detail classes failed', error)
  }
}

async function loadOneToOneOptions(searchKey = '', reset = true) {
  if (reset) {
    oneToOnePagination.value.current = 1
    oneToOneFinished.value = false
  }
  oneToOneSearchKey.value = searchKey
  try {
    const res = await getOneToOneListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: oneToOnePagination.value.pageSize,
        pageIndex: oneToOnePagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey: searchKey || undefined,
      },
    })
    if (res.code !== 200)
      return
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    const resultData = list.map(item => ({
      id: String(item.id ?? ''),
      value: String(item.name || item.studentName || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    oneToOneOptions.value = reset
      ? mergeFilterOptions(oneToOneOptions.value, resultData, filterOneToOneId.value)
      : mergeFilterOptions(oneToOneOptions.value, [...oneToOneOptions.value, ...resultData], filterOneToOneId.value)
    oneToOnePagination.value.total = Number(res.result?.total || resultData.length || 0)
    oneToOneFinished.value = oneToOneOptions.value.length >= oneToOnePagination.value.total
  }
  catch (error) {
    console.error('load class comment detail one to one failed', error)
  }
}

async function loadTeacherOptions(searchKey = '', reset = true) {
  if (reset) {
    teacherPagination.value.current = 1
    teacherFinished.value = false
  }
  teacherSearchKey.value = searchKey
  try {
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: teacherPagination.value.pageSize,
        pageIndex: teacherPagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey: searchKey || undefined,
      },
      sortModel: {},
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.nickName || item.name || item.value || item.id || '').trim(),
      mobile: String(item.mobile || item.phone || '').trim() || undefined,
    })).filter(item => item.id && item.value)
    teacherOptions.value = reset
      ? mergeFilterOptions(teacherOptions.value, resultData, filterTeacherIds.value)
      : mergeFilterOptions(teacherOptions.value, [...teacherOptions.value, ...resultData], filterTeacherIds.value)
    teacherPagination.value.total = Number(res.total || resultData.length || 0)
    teacherFinished.value = teacherOptions.value.length >= teacherPagination.value.total
  }
  catch (error) {
    console.error('load class comment detail teachers failed', error)
  }
}

async function loadAssistantTeacherOptions(searchKey = '', reset = true) {
  if (reset) {
    assistantTeacherPagination.value.current = 1
    assistantTeacherFinished.value = false
  }
  assistantTeacherSearchKey.value = searchKey
  try {
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: assistantTeacherPagination.value.pageSize,
        pageIndex: assistantTeacherPagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey: searchKey || undefined,
      },
      sortModel: {},
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result) ? res.result : []).map(item => ({
      id: String(item.id ?? ''),
      value: String(item.nickName || item.name || item.value || item.id || '').trim(),
      mobile: String(item.mobile || item.phone || '').trim() || undefined,
    })).filter(item => item.id && item.value)
    assistantTeacherOptions.value = reset
      ? mergeFilterOptions(assistantTeacherOptions.value, resultData, filterAssistantTeacherIds.value)
      : mergeFilterOptions(assistantTeacherOptions.value, [...assistantTeacherOptions.value, ...resultData], filterAssistantTeacherIds.value)
    assistantTeacherPagination.value.total = Number(res.total || resultData.length || 0)
    assistantTeacherFinished.value = assistantTeacherOptions.value.length >= assistantTeacherPagination.value.total
  }
  catch (error) {
    console.error('load class comment detail assistants failed', error)
  }
}

async function loadList() {
  loading.value = true
  try {
    const res = await getClassCommentStudentPagedListApi({
      queryModel: buildQueryModel(),
      pageRequestModel: {
        needTotal: true,
        pageIndex: pagination.value.current,
        pageSize: pagination.value.pageSize,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        startTime: sortStartTime.value,
      },
    })
    if (res.code !== 200)
      throw new Error(res.message || '获取课堂点评明细失败')
    dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
    pagination.value.total = Number(res.result?.total || 0)
  }
  catch (error: any) {
    console.error('load class comment detail failed', error)
    dataSource.value = []
    pagination.value.total = 0
    messageService.error(error?.response?.data?.message || error?.message || '获取课堂点评明细失败')
  }
  finally {
    loading.value = false
  }
}

function handleScheduleDateFilter(value: unknown) {
  if (!Array.isArray(value) || value.length < 2)
    return
  const start = dayjs(String(value[0] || ''))
  const end = dayjs(String(value[1] || ''))
  if (!start.isValid() || !end.isValid())
    return
  if (
    filterDateRange.value?.[0]?.format('YYYY-MM-DD') === start.format('YYYY-MM-DD')
    && filterDateRange.value?.[1]?.format('YYYY-MM-DD') === end.format('YYYY-MM-DD')
  ) {
    return
  }
  filterDateRange.value = [start, end]
}

function handleLessonFilter(value: unknown) {
  filterLessonId.value = normalizeFilterValue(value)
}

function handleTeacherFilter(value: unknown) {
  filterTeacherIds.value = normalizeFilterValues(value)
}

function handleAssistantTeacherFilter(value: unknown) {
  filterAssistantTeacherIds.value = normalizeFilterValues(value)
}

function handleClassFilter(value: unknown) {
  filterClassId.value = normalizeFilterValue(value)
}

function handleOneToOneFilter(value: unknown) {
  filterOneToOneId.value = normalizeFilterValue(value)
}

function handleScheduleTypeFilter(value: unknown) {
  filterScheduleTypes.value = normalizeFilterValues(value)
}

function handleCommentStatusFilter(value: unknown) {
  filterCommentStatus.value = normalizeBooleanFilter(value)
}

function handleReadStatusFilter(value: unknown) {
  filterReadStatus.value = normalizeBooleanFilter(value)
}

function handleParentFeedbackStatusFilter(value: unknown) {
  filterParentFeedbackStatus.value = normalizeBooleanFilter(value)
}

function handleStudentFilter(value: unknown) {
  filterStudentId.value = normalizeFilterValue(value)
}

function handleTableChange(page: { current?: number, pageSize?: number }, _filters: any, sorter: any) {
  pagination.value.current = Number(page?.current || 1)
  pagination.value.pageSize = Number(page?.pageSize || pagination.value.pageSize)
  if (sorter?.order === 'ascend')
    sortStartTime.value = 1
  else
    sortStartTime.value = 2
  loadList()
}

watch(
  [
    filterDateRange,
    filterLessonId,
    filterTeacherIds,
    filterAssistantTeacherIds,
    filterClassId,
    filterOneToOneId,
    filterScheduleTypes,
    filterCommentStatus,
    filterReadStatus,
    filterParentFeedbackStatus,
    filterStudentId,
  ],
  () => {
    pagination.value.current = 1
    loadList()
  },
  { deep: true },
)

onMounted(() => {
  loadList()
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        :display-array="displayArray"
        :default-schedule-date-vals="defaultScheduleDateVals"
        :locked-condition-types="['scheduleDate']"
        :whole-condition-clear-types="['scheduleTeacher', 'assistantTeacher', 'scheduleType']"
        :schedule-date-disable-future="true"
        :schedule-course-options="courseOptions"
        :schedule-course-finished="courseFinished"
        :on-schedule-course-dropdown-visible-change="() => loadCourseOptions()"
        :on-schedule-course-search="loadCourseOptions"
        :schedule-teacher-options="teacherOptions"
        :schedule-teacher-finished="teacherFinished"
        :on-schedule-teacher-dropdown-visible-change="() => loadTeacherOptions('', true)"
        :on-schedule-teacher-search="keyword => loadTeacherOptions(keyword, true)"
        :on-schedule-teacher-load-more="() => { teacherPagination.current += 1; return loadTeacherOptions(teacherSearchKey, false) }"
        :assistant-teacher-options="assistantTeacherOptions"
        :assistant-teacher-finished="assistantTeacherFinished"
        :on-assistant-teacher-dropdown-visible-change="() => loadAssistantTeacherOptions('', true)"
        :on-assistant-teacher-search="keyword => loadAssistantTeacherOptions(keyword, true)"
        :on-assistant-teacher-load-more="() => { assistantTeacherPagination.current += 1; return loadAssistantTeacherOptions(assistantTeacherSearchKey, false) }"
        :schedule-class-options="classOptions"
        :schedule-class-finished="classFinished"
        :on-schedule-class-dropdown-visible-change="() => loadClassOptions('', true)"
        :on-schedule-class-search="keyword => loadClassOptions(keyword, true)"
        :on-schedule-class-load-more="() => { classPagination.current += 1; return loadClassOptions(classSearchKey, false) }"
        :schedule-one-to-one-options="oneToOneOptions"
        :schedule-one-to-one-finished="oneToOneFinished"
        :on-schedule-one-to-one-dropdown-visible-change="() => loadOneToOneOptions('', true)"
        :on-schedule-one-to-one-search="keyword => loadOneToOneOptions(keyword, true)"
        :on-schedule-one-to-one-load-more="() => { oneToOnePagination.current += 1; return loadOneToOneOptions(oneToOneSearchKey, false) }"
        :schedule-course-label="'课程'"
        :schedule-class-label="'班级'"
        :schedule-one-to-one-label="'1对1'"
        :schedule-type-label="'类型'"
        :schedule-type-options="scheduleTypeOptions"
        :comment-status-label="'是否点评'"
        :comment-status-options="commentStatusOptions"
        :read-status-label="'已读/未读'"
        :read-status-options="readStatusOptions"
        :parent-feedback-status-label="'课评反馈'"
        :parent-feedback-status-options="parentFeedbackStatusOptions"
        :is-quick-show="false"
        @update:schedule-date-filter="handleScheduleDateFilter"
        @update:schedule-course-filter="handleLessonFilter"
        @update:schedule-teacher-filter="handleTeacherFilter"
        @update:assistant-teacher-filter="handleAssistantTeacherFilter"
        @update:schedule-class-filter="handleClassFilter"
        @update:schedule-one-to-one-filter="handleOneToOneFilter"
        @update:schedule-type-filter="handleScheduleTypeFilter"
        @update:comment-status-filter="handleCommentStatusFilter"
        @update:read-status-filter="handleReadStatusFilter"
        @update:parent-feedback-status-filter="handleParentFeedbackStatusFilter"
        @update:stu-phone-search-filter="handleStudentFilter"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ pagination.total }} 条数据
          </div>
          <div class="edit flex" />
        </div>
        <div class="table-content mt-2">
          <a-table
            row-key="studentTeachingRecordId"
            :loading="loading"
            :data-source="dataSource"
            :pagination="tablePagination"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'classDateTime'">
                <div class="text-#222">
                  {{ formatDateTime(record).dateText }}
                </div>
                <div class="text-#222">
                  {{ formatDateTime(record).timeText }}
                </div>
              </template>
              <template v-if="column.key === 'type'">
                <div class="justify-between flex-center">
                  <span>{{ sourceTypeText(record) }}</span>
                  <img
                    v-if="sourceTypeImage(record)"
                    class="type-tag-image"
                    height="45"
                    :src="sourceTypeImage(record)"
                    alt=""
                  >
                </div>
              </template>
              <template v-if="column.key === 'linkClassOr1v1'">
                <div class="text-#222">
                  {{ record.sourceName || '-' }}
                </div>
              </template>
              <template v-if="column.key === 'linkCourse'">
                <div class="text-#222">
                  {{ record.lessonName || '-' }}
                </div>
              </template>
              <template v-if="column.key === 'teacher'">
                <div class="text-#222">
                  {{ record.teacherName || '-' }}
                </div>
              </template>
              <template v-if="column.key === 'student'">
                <StudentAvatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :phone="record.studentPhone || ''"
                  :avatar-url="record.avatar || ''"
                  :show-age="false"
                  :show-gender="false"
                />
              </template>
              <template v-if="column.key === 'isComment'">
                <span
                  :class="record.isComment ? 'bg-#e6f4ff text-#1677ff' : 'bg-#fff7e6 text-#fa8c16'"
                  class="text-3 px2 py1 rounded-10"
                >
                  {{ record.isComment ? '已点评' : '未点评' }}
                </span>
              </template>
              <template v-if="column.key === 'readStatus'">
                <div class="text-#222">
                  {{ readStatusText(record) }}
                </div>
              </template>
              <template v-if="column.key === 'parentFeedback'">
                <div class="text-#222">
                  {{ parentFeedbackText(record) }}
                </div>
              </template>
              <template v-if="column.key === 'parentFeedbackGrade'">
                <div class="text-#222">
                  {{ parentFeedbackGradeText(record) }}
                </div>
              </template>
              <template v-if="column.key === 'parentFeedbackContent'">
                <div class="text-#222">
                  {{ displayText(record.parentFeedbackContent) }}
                </div>
              </template>
              <template v-if="column.key === 'subTeacher'">
                <div class="text-#222">
                  {{ displayText(record.assistants) }}
                </div>
              </template>
              <template v-if="column.key === 'classRoom'">
                <div class="text-#222">
                  {{ displayText(record.classRoomName) }}
                </div>
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="14">
                  <a class="font500" @click="handleViewPending">去点评</a>
                  <a class="font500" @click="handleViewPending">查看</a>
                </a-space>
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

.tip {
  padding: 10px 24px 10px 14px;
  background: #e6f0ff;
}

.type-tag-image {
  opacity: 0.4;
}

.upNew {
  position: relative;

  &::before {
    position: absolute;
    top: -12px;
    left: -22px;
    z-index: 999;
    width: 39px;
    height: 22px;
    background: url(data:image/png;base64,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);
    background-size: contain;
    content: "";
  }
}
</style>
