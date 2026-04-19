<script setup lang="ts">
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { computed, onMounted, ref, watch } from 'vue'
import scheduleClassImage from '@/assets/images/timetable/schedule-class.png'
import scheduleOneToOneImage from '@/assets/images/timetable/schedule-one2one.png'
import {
  getClassCommentPagedListApi,
  type ClassCommentItem,
} from '@/api/edu-center/class-record'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import ClassReviewDrawer from '@/components/home-center/class-comment/classReviewDrawer.vue'
import { useTableColumns } from '@/composables/useTableColumns'
import messageService from '@/utils/messageService'

interface FilterOption {
  id: string
  value: string
}

const displayArray = ref([
  'scheduleCourse',
  'scheduleTeacher',
  'scheduleClass',
  'scheduleOneToOne',
  'scheduleDate',
  'scheduleType',
])

const scheduleTypeOptions = [
  { id: '1', value: '班级' },
  { id: '2', value: '1对1' },
]

const loading = ref(false)
const dataSource = ref<ClassCommentItem[]>([])
const reviewDrawerOpen = ref(false)
const currentReviewRecord = ref<Partial<ClassCommentItem> | null>(null)
const currentReviewTab = ref<'0' | '1'>('1')
const sortStartTime = ref(2)

const filterDateRange = ref<[Dayjs, Dayjs] | null>(null)
const filterLessonId = ref<string | undefined>(undefined)
const filterTeacherIds = ref<string[]>([])
const filterClassId = ref<string | undefined>(undefined)
const filterOneToOneId = ref<string | undefined>(undefined)
const filterScheduleTypes = ref<string[]>([])

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

const classPagination = ref({ current: 1, pageSize: 20, total: 0 })
const oneToOnePagination = ref({ current: 1, pageSize: 20, total: 0 })
const teacherPagination = ref({ current: 1, pageSize: 20, total: 0 })

const classSearchKey = ref('')
const oneToOneSearchKey = ref('')
const teacherSearchKey = ref('')

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
    title: '记录统计',
    key: 'commentStatistics',
    dataIndex: 'commentStatistics',
    width: 110,
  },
  {
    title: '已读/未读',
    dataIndex: 'readOrUnread',
    key: 'readOrUnread',
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
  storageKey: 'class-comment-list',
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

function formatDateTime(record: Partial<ClassCommentItem>) {
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

function sourceTypeText(record: Partial<ClassCommentItem>) {
  const sourceType = Number(record.sourceType || 0)
  if (sourceType === 2)
    return '1对1'
  if (sourceType === 3)
    return '试听'
  return '班级'
}

function sourceTypeImage(record: Partial<ClassCommentItem>) {
  const sourceType = Number(record.sourceType || 0)
  if (sourceType === 2)
    return scheduleOneToOneImage
  if (sourceType === 1)
    return scheduleClassImage
  return ''
}

function commentStatisticsText(record: Partial<ClassCommentItem>) {
  const commentCount = Number(record.commentCount || 0)
  const unCommentCount = Number(record.unCommentCount || 0)
  return `${commentCount}/${commentCount + unCommentCount}`
}

function isFullyCommented(record: Partial<ClassCommentItem>) {
  const commentCount = Number(record.commentCount || 0)
  const unCommentCount = Number(record.unCommentCount || 0)
  return commentCount + unCommentCount > 0 && unCommentCount === 0
}

function hasReadStatistics(record: Partial<ClassCommentItem>) {
  return Number(record.readCount || 0) + Number(record.unReadCount || 0) > 0 || Number(record.commentCount || 0) > 0
}

function displayReadCount(record: Partial<ClassCommentItem>) {
  return Number(record.readCount || 0)
}

function displayUnreadCount(record: Partial<ClassCommentItem>) {
  const readCount = displayReadCount(record)
  const unReadCount = Number(record.unReadCount || 0)
  const commentCount = Number(record.commentCount || 0)
  if (readCount + unReadCount > 0)
    return unReadCount
  return commentCount
}

function openReviewDrawer(record: Partial<ClassCommentItem> | undefined, initialActiveKey: '0' | '1') {
  currentReviewRecord.value = record ? { ...record } : null
  currentReviewTab.value = initialActiveKey
  reviewDrawerOpen.value = true
}

function handleOpenReviewDrawer(record?: Partial<ClassCommentItem>) {
  openReviewDrawer(record, '1')
}

function handleViewReviewDrawer(record?: Partial<ClassCommentItem>) {
  openReviewDrawer(record, '0')
}

function buildQueryModel() {
  return {
    teachingStartTime: filterDateRange.value?.[0]?.format('YYYY-MM-DD'),
    teachingEndTime: filterDateRange.value?.[1]?.format('YYYY-MM-DD'),
    teachingRecordTypes: filterScheduleTypes.value.length
      ? filterScheduleTypes.value.map(item => Number(item)).filter(Boolean)
      : [1, 2],
    lessonId: filterLessonId.value,
    teacherIds: filterTeacherIds.value.length ? filterTeacherIds.value : undefined,
    classId: filterClassId.value,
    one2OneId: filterOneToOneId.value,
    classTeacherIds: [],
    one2OneTeacherIds: [],
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
    console.error('load class comment courses failed', error)
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
    console.error('load class comment classes failed', error)
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
    console.error('load class comment one to one failed', error)
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
    console.error('load class comment teachers failed', error)
  }
}

async function loadList() {
  loading.value = true
  try {
    const res = await getClassCommentPagedListApi({
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
      throw new Error(res.message || '获取康复记录列表失败')
    dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
    pagination.value.total = Number(res.result?.total || 0)
  }
  catch (error: any) {
    console.error('load class comment list failed', error)
    dataSource.value = []
    pagination.value.total = 0
    messageService.error(error?.response?.data?.message || error?.message || '获取康复记录列表失败')
  }
  finally {
    loading.value = false
  }
}

function handleScheduleDateFilter(value: unknown) {
  if (!Array.isArray(value) || value.length < 2) {
    filterDateRange.value = null
    return
  }
  const start = dayjs(String(value[0] || ''))
  const end = dayjs(String(value[1] || ''))
  if (!start.isValid() || !end.isValid()) {
    filterDateRange.value = null
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

function handleClassFilter(value: unknown) {
  filterClassId.value = normalizeFilterValue(value)
}

function handleOneToOneFilter(value: unknown) {
  filterOneToOneId.value = normalizeFilterValue(value)
}

function handleScheduleTypeFilter(value: unknown) {
  filterScheduleTypes.value = Array.isArray(value) ? value.map(item => String(item || '')).filter(Boolean) : []
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
    filterClassId,
    filterOneToOneId,
    filterScheduleTypes,
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
        :whole-condition-clear-types="['scheduleTeacher', 'scheduleType']"
        :is-quick-show="false"
        @update:schedule-date-filter="handleScheduleDateFilter"
        @update:schedule-course-filter="handleLessonFilter"
        @update:schedule-teacher-filter="handleTeacherFilter"
        @update:schedule-class-filter="handleClassFilter"
        @update:schedule-one-to-one-filter="handleOneToOneFilter"
        @update:schedule-type-filter="handleScheduleTypeFilter"
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
            row-key="teachingRecordId"
            :loading="loading"
            :data-source="dataSource"
            :pagination="tablePagination"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            size="small"
            @change="handleTableChange"
          >
            <template #headerCell="{ column }">
              <template v-if="column.key === 'commentStatistics'">
                <span class="mr-1">{{ column.title }}</span>
                <a-tooltip color="#666">
                  <template #title>
                    已记录人数/应记录人数
                  </template>
                  <ExclamationCircleOutlined />
                </a-tooltip>
              </template>
            </template>
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
              <template v-if="column.key === 'commentStatistics'">
                <div class="text-#222">
                  {{ commentStatisticsText(record) }}
                </div>
              </template>
              <template v-if="column.key === 'readOrUnread'">
                <template v-if="hasReadStatistics(record)">
                  <div class="text-#222">
                    已读{{ displayReadCount(record) }}人
                  </div>
                  <div class="text-#888 text-3">
                    未读{{ displayUnreadCount(record) }}人
                  </div>
                </template>
                <div v-else class="text-#222">
                  -
                </div>
              </template>
              <template v-if="column.key === 'subTeacher'">
                <div class="text-#222">
                  {{ record.assistants || '-' }}
                </div>
              </template>
              <template v-if="column.key === 'classRoom'">
                <div class="text-#222">
                  {{ record.classRoomName || '-' }}
                </div>
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="14">
                  <a v-if="!isFullyCommented(record)" class="font500" @click="handleOpenReviewDrawer(record)">去记录</a>
                  <a class="font500" @click="handleViewReviewDrawer(record)">查看</a>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <ClassReviewDrawer
      v-model="reviewDrawerOpen"
      :initial-active-key="currentReviewTab"
      :record="currentReviewRecord"
      @updated="loadList"
    />
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
