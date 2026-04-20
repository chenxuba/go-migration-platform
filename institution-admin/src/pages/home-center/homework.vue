<script setup lang="ts">
import { QuestionCircleOutlined } from '@ant-design/icons-vue'
import { computed, h, nextTick, onMounted, reactive, ref } from 'vue'
import { Modal, Tooltip } from 'ant-design-vue'
import dayjs from 'dayjs'
import afterSchoolTasksModel from './components/afterSchoolTasksModel.vue'
import { pageGroupClassSelectionApi } from '@/api/edu-center/group-class'
import { pageOneToOneSelectionApi } from '@/api/edu-center/one-to-one'
import {
  deleteHomeworkApi,
  homeworkStatisticsApi,
  pageHomeworksApi,
  type HomeworkListItem,
} from '@/api/home-center/homework'
import { useTableColumns } from '@/composables/useTableColumns'
import messageService from '@/utils/messageService'

interface FilterOption {
  id: string
  value: string
}

const allFilterRef = ref()
const loading = ref(false)
const tableData = ref<HomeworkListItem[]>([])

const modalOpen = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const currentHomeworkId = ref('')

const quickCounts = reactive({
  unevaluatedCount: 0,
  unsubmittedCount: 0,
})

const displayArray = ref([
  'scheduleClass',
  'scheduleOneToOne',
  'createUser',
  'applyTime',
  'classEndingTime',
])

function getDefaultPublishRange() {
  return [
    dayjs().startOf('month').format('YYYY-MM-DD'),
    dayjs().format('YYYY-MM-DD'),
  ]
}

const queryState = reactive({
  classId: undefined as string | undefined,
  one2OneId: undefined as string | undefined,
  teacherIds: undefined as string[] | undefined,
  publishRange: getDefaultPublishRange(),
  endRange: [] as string[],
  hasUnevaluated: undefined as boolean | undefined,
  hasUnsubmitted: undefined as boolean | undefined,
})

const sortState = reactive({
  publishTime: 0,
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  pageSizeOptions: ['10', '20', '50', '100'],
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
})

const scheduleClassOptions = ref<FilterOption[]>([])
const scheduleOneToOneOptions = ref<FilterOption[]>([])
const scheduleClassFinished = ref(false)
const scheduleOneToOneFinished = ref(false)
const scheduleClassPagination = reactive({ current: 1, pageSize: 50, total: 0 })
const scheduleOneToOnePagination = reactive({ current: 1, pageSize: 50, total: 0 })
const scheduleClassSearchKey = ref('')
const scheduleOneToOneSearchKey = ref('')

const customQuickFilters = computed(() => [
  { id: 1, name: '待全部点评', count: quickCounts.unevaluatedCount },
  { id: 2, name: '待全部提交', count: quickCounts.unsubmittedCount },
])

const allColumns = ref([
  {
    title: '任务名称（班级/1对1）',
    dataIndex: 'homeworkName',
    key: 'homeworkName',
    width: 240,
  },
  {
    title: '发布内容',
    dataIndex: 'publishContent',
    key: 'publishContent',
    width: 260,
  },
  {
    title: () => h('span', { class: 'homework-column-title' }, [
      '提交任务率',
      h(Tooltip, { title: '已交学员数/学员总数' }, {
        default: () => h(QuestionCircleOutlined, { class: 'homework-column-title__icon' }),
      }),
    ]),
    dataIndex: 'submitRate',
    key: 'submitRate',
    width: 150,
  },
  {
    title: '待点评数量',
    dataIndex: 'pendingCorrectionNum',
    key: 'pendingCorrectionNum',
    width: 130,
  },
  {
    title: '未读',
    dataIndex: 'unreadCount',
    key: 'unreadCount',
    width: 100,
  },
  {
    title: '发布人',
    dataIndex: 'publishUser',
    key: 'publishUser',
    width: 120,
  },
  {
    title: '发布时间',
    dataIndex: 'publishTime',
    key: 'publishTime',
    width: 160,
    sorter: true,
    defaultSortOrder: 'descend',
  },
  {
    title: '截止时间',
    dataIndex: 'deadlineTime',
    key: 'deadlineTime',
    width: 160,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 120,
  },
])

const { filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'home-center-homework',
  allColumns,
  excludeKeys: ['action'],
})

const WEEKDAY_LABELS = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']

function normalizeSingleValue(value: unknown) {
  if (Array.isArray(value))
    return value.length ? String(value[0] || '').trim() || undefined : undefined
  const text = String(value ?? '').trim()
  return text || undefined
}

function normalizeDateRange(value: unknown) {
  if (!Array.isArray(value) || value.length !== 2)
    return []
  const list = value.map(item => String(item || '').trim()).filter(Boolean)
  return list.length === 2 ? list : []
}

function mergeFilterOptions(previous: FilterOption[], incoming: FilterOption[], selectedValues: string | string[] | undefined = []) {
  const selectedSet = new Set((Array.isArray(selectedValues) ? selectedValues : [selectedValues]).map(value => String(value || '')).filter(Boolean))
  const optionMap = new Map<string, FilterOption>()
  previous.forEach((item) => {
    if (selectedSet.has(item.id))
      optionMap.set(item.id, item)
  })
  incoming.forEach((item) => {
    if (item.id)
      optionMap.set(item.id, item)
  })
  return [...optionMap.values()]
}

function resetQueryState() {
  queryState.classId = undefined
  queryState.one2OneId = undefined
  queryState.teacherIds = undefined
  queryState.publishRange = getDefaultPublishRange()
  queryState.endRange = []
  queryState.hasUnevaluated = undefined
  queryState.hasUnsubmitted = undefined
}

function syncDefaultPublishRangeFilter() {
  allFilterRef.value?.setApplyTimeFilter?.(getDefaultPublishRange(), false)
}

function buildQueryModel(includeQuick = true) {
  return {
    classId: queryState.classId,
    one2OneId: queryState.one2OneId,
    teacherIds: queryState.teacherIds,
    publishStartTime: queryState.publishRange[0],
    publishEndTime: queryState.publishRange[1],
    endStartTime: queryState.endRange[0],
    endEndTime: queryState.endRange[1],
    hasUnevaluated: includeQuick ? queryState.hasUnevaluated : undefined,
    hasUnsubmitted: includeQuick ? queryState.hasUnsubmitted : undefined,
  }
}

async function fetchHomeworkList(id?: string | number, type?: string) {
  loading.value = true
  try {
    const res = await pageHomeworksApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.pageSize,
        pageIndex: pagination.current,
        skipCount: (pagination.current - 1) * pagination.pageSize,
      },
      sortModel: {
        publishTime: sortState.publishTime,
      },
      queryModel: buildQueryModel(true),
    })
    if (res.code !== 200) {
      messageService.error(res.message || '获取课后任务列表失败')
      return
    }
    tableData.value = Array.isArray(res.result?.list) ? res.result.list : []
    pagination.total = Number(res.result?.total || 0)
    allFilterRef.value?.clearQuickFilter?.(id, type)
  }
  catch (error) {
    console.error('fetch homework list failed', error)
    messageService.error('获取课后任务列表失败')
  }
  finally {
    loading.value = false
  }
}

async function fetchHomeworkStatistics() {
  try {
    const res = await homeworkStatisticsApi(buildQueryModel(false))
    if (res.code !== 200) {
      quickCounts.unevaluatedCount = 0
      quickCounts.unsubmittedCount = 0
      return
    }
    quickCounts.unevaluatedCount = Number(res.result?.unevaluatedCount || 0)
    quickCounts.unsubmittedCount = Number(res.result?.unsubmittedCount || 0)
  }
  catch (error) {
    console.error('fetch homework statistics failed', error)
    quickCounts.unevaluatedCount = 0
    quickCounts.unsubmittedCount = 0
  }
}

async function refreshData(id?: string | number, type?: string) {
  await Promise.all([fetchHomeworkList(id, type), fetchHomeworkStatistics()])
}

async function loadScheduleClassOptions(searchKey = '', reset = true) {
  if (reset) {
    scheduleClassPagination.current = 1
    scheduleClassFinished.value = false
  }
  scheduleClassSearchKey.value = searchKey
  try {
    const res = await pageGroupClassSelectionApi({
      queryModel: {
        className: searchKey || undefined,
        status: [1],
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: scheduleClassPagination.pageSize,
        pageIndex: scheduleClassPagination.current,
        skipCount: 0,
      },
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result?.list) ? res.result.list : []).map(item => ({
      id: String(item.id || ''),
      value: String(item.name || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleClassOptions.value = reset
      ? mergeFilterOptions(scheduleClassOptions.value, resultData, queryState.classId)
      : mergeFilterOptions(scheduleClassOptions.value, [...scheduleClassOptions.value, ...resultData], queryState.classId)
    scheduleClassPagination.total = Number(res.result?.total || resultData.length || 0)
    scheduleClassFinished.value = scheduleClassOptions.value.length >= scheduleClassPagination.total
  }
  catch (error) {
    console.error('load homework class options failed', error)
  }
}

async function loadScheduleOneToOneOptions(searchKey = '', reset = true) {
  if (reset) {
    scheduleOneToOnePagination.current = 1
    scheduleOneToOneFinished.value = false
  }
  scheduleOneToOneSearchKey.value = searchKey
  try {
    const res = await pageOneToOneSelectionApi({
      queryModel: {
        searchKey: searchKey || undefined,
        status: [1],
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: scheduleOneToOnePagination.pageSize,
        pageIndex: scheduleOneToOnePagination.current,
        skipCount: 0,
      },
    })
    if (res.code !== 200)
      return
    const resultData = (Array.isArray(res.result?.list) ? res.result.list : []).map(item => ({
      id: String(item.id || ''),
      value: String(item.name || `${item.studentName || ''} - ${item.lessonName || ''}` || item.id || '').trim(),
    })).filter(item => item.id && item.value)
    scheduleOneToOneOptions.value = reset
      ? mergeFilterOptions(scheduleOneToOneOptions.value, resultData, queryState.one2OneId)
      : mergeFilterOptions(scheduleOneToOneOptions.value, [...scheduleOneToOneOptions.value, ...resultData], queryState.one2OneId)
    scheduleOneToOnePagination.total = Number(res.result?.total || resultData.length || 0)
    scheduleOneToOneFinished.value = scheduleOneToOneOptions.value.length >= scheduleOneToOnePagination.total
  }
  catch (error) {
    console.error('load homework one-to-one options failed', error)
  }
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
  scheduleClassPagination.current += 1
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
  scheduleOneToOnePagination.current += 1
  await loadScheduleOneToOneOptions(scheduleOneToOneSearchKey.value, false)
}

function handleFilterUpdate(updates: Partial<typeof queryState> = {}, isClearAll = false, id?: string | number, type?: string) {
  if (isClearAll) {
    resetQueryState()
    nextTick(() => {
      syncDefaultPublishRangeFilter()
    })
  }
  else {
    Object.assign(queryState, updates)
  }
  pagination.current = 1
  void refreshData(id, type)
}

function handleQuickFilterChange(value: unknown, isClearAll?: boolean, id?: string | number, type?: string) {
  if (isClearAll) {
    handleFilterUpdate({}, true)
    return
  }
  if (value === 'unevaluated') {
    handleFilterUpdate({
      hasUnevaluated: true,
      hasUnsubmitted: undefined,
    }, false, id, type)
    return
  }
  if (value === 'unsubmitted') {
    handleFilterUpdate({
      hasUnevaluated: undefined,
      hasUnsubmitted: true,
    }, false, id, type)
    return
  }
  handleFilterUpdate({
    hasUnevaluated: undefined,
    hasUnsubmitted: undefined,
  }, false, id, type)
}

function handleScheduleClassFilter(value: unknown, isClearAll?: boolean) {
  handleFilterUpdate({ classId: normalizeSingleValue(value) }, isClearAll)
}

function handleScheduleOneToOneFilter(value: unknown, isClearAll?: boolean) {
  handleFilterUpdate({ one2OneId: normalizeSingleValue(value) }, isClearAll)
}

function handleCreateUserFilter(value: unknown, isClearAll?: boolean, id?: string | number, type?: string) {
  const current = normalizeSingleValue(value)
  handleFilterUpdate({ teacherIds: current ? [current] : undefined }, isClearAll, id, type)
}

function handlePublishTimeFilter(value: unknown, isClearAll?: boolean) {
  handleFilterUpdate({ publishRange: normalizeDateRange(value) }, isClearAll)
}

function handleDeadlineTimeFilter(value: unknown, isClearAll?: boolean) {
  handleFilterUpdate({ endRange: normalizeDateRange(value) }, isClearAll)
}

function handleTableChange(pageInfo: { current?: number, pageSize?: number }, _filters: unknown, sorter: { order?: string } | Array<{ order?: string }>) {
  pagination.current = Number(pageInfo.current || pagination.current)
  pagination.pageSize = Number(pageInfo.pageSize || pagination.pageSize)
  if (!Array.isArray(sorter))
    sortState.publishTime = sorter?.order === 'ascend' ? 1 : 0
  void fetchHomeworkList()
}

function hasDateTimeValue(value?: string) {
  if (!value)
    return false
  return dayjs(value).isValid()
}

function isFutureDateTime(value?: string) {
  if (!value)
    return false
  const current = dayjs(value)
  if (!current.isValid())
    return false
  return current.isAfter(dayjs())
}

function formatDateText(value?: string) {
  if (!value)
    return '--'
  const current = dayjs(value)
  if (!current.isValid())
    return '--'
  return `${current.format('YYYY-MM-DD')}（${WEEKDAY_LABELS[current.day()] || '--'}）`
}

function formatTimeText(value?: string) {
  if (!value)
    return '--'
  const current = dayjs(value)
  if (!current.isValid())
    return '--'
  return current.format('HH:mm')
}

function formatSubmitRate(record: Partial<HomeworkListItem> & Record<string, any>) {
  const total = Number(record.studentCount || 0)
  const submitted = Number(record.submittedCount || 0)
  if (total <= 0)
    return '0%'
  return `${Math.round((submitted / total) * 100)}%`
}

function openCreateModal() {
  modalMode.value = 'create'
  currentHomeworkId.value = ''
  modalOpen.value = true
}

function openEditModal(record: Partial<HomeworkListItem> & Record<string, any>) {
  modalMode.value = 'edit'
  currentHomeworkId.value = String(record.id || '')
  modalOpen.value = true
}

function handleModalSuccess() {
  modalOpen.value = false
  void refreshData()
}

function handleDelete(record: Partial<HomeworkListItem> & Record<string, any>) {
  Modal.confirm({
    title: '删除课后任务',
    content: `删除后不可恢复，确认删除「${record.title || '该任务'}」吗？`,
    okText: '确认删除',
    cancelText: '取消',
    async onOk() {
      const res = await deleteHomeworkApi({ id: String(record.id || '') })
      if (res.code !== 200) {
        messageService.error(res.message || '删除课后任务失败')
        return
      }
      messageService.success('删除成功')
      await refreshData()
    },
  })
}

onMounted(() => {
  nextTick(() => {
    syncDefaultPublishRangeFilter()
  })
  void refreshData()
})
</script>

<template>
  <div class="homework-page">
    <div class="filter-wrap bg-white rounded-4 px-4 py-3">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :custom-quick-filters="customQuickFilters"
        :custom-quick-filter-values="{ 1: 'unevaluated', 2: 'unsubmitted' }"
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
        create-user-label="发布人"
        create-user-placeholder="请输入发布人"
        apply-time-label="发布时间"
        class-ending-time-label="截止时间"
        schedule-class-label="班级"
        schedule-one-to-one-label="1对1"
        is-quick-show
        @update:quick-filter="handleQuickFilterChange"
        @update:schedule-class-filter="handleScheduleClassFilter"
        @update:schedule-one-to-one-filter="handleScheduleOneToOneFilter"
        @update:create-user-filter="handleCreateUserFilter"
        @update:apply-time-filter="handlePublishTimeFilter"
        @update:class-ending-time-filter="handleDeadlineTimeFilter"
      />
    </div>

    <div class="homework-panel bg-white rounded-4 mt-3 px-5 py-4">
      <div class="homework-panel__header">
        <div class="homework-panel__summary">
          <div class="homework-panel__title">
            共 {{ pagination.total }} 个课后任务
          </div>
        </div>

        <a-button type="primary" @click="openCreateModal">
          新建课后任务
        </a-button>
      </div>

      <a-table
        row-key="id"
        class="mt-4"
        size="small"
        :loading="loading"
        :data-source="tableData"
        :columns="filteredColumns"
        :pagination="pagination"
        :scroll="{ x: totalWidth }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'homeworkName'">
            <div class="homework-name-cell">
              <div class="homework-name-cell__title">
                {{ record.title || '-' }}
              </div>
              <div class="homework-name-cell__meta">
                <span class="homework-source-tag" :class="record.sourceType === 2 ? 'is-one-to-one' : ''">
                  {{ record.sourceType === 2 ? '1对1' : '班级' }}
                </span>
                <span class="homework-name-cell__source">{{ record.sourceName || '-' }}</span>
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'publishContent'">
            <div class="homework-content-cell">
              <clamped-text :lines="2" :text="record.content || '-'" />
            </div>
          </template>

          <template v-else-if="column.key === 'submitRate'">
            <div class="homework-metric-cell">
              <div class="homework-metric-cell__value">
                {{ formatSubmitRate(record) }}
              </div>
              <div class="homework-metric-cell__desc">
                已交{{ record.submittedCount || 0 }}人 / 应交{{ record.studentCount || 0 }}人
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'pendingCorrectionNum'">
            <span :class="record.unevaluatedCount > 0 ? 'text-#fa8c16 font-500' : 'text-#8c8c8c'">
              {{ record.unevaluatedCount || 0 }} 人
            </span>
          </template>

          <template v-else-if="column.key === 'unreadCount'">
            <span :class="record.unreadCount > 0 ? 'text-#ff4d4f font-500' : 'text-#8c8c8c'">
              {{ record.unreadCount || 0 }} 人
            </span>
          </template>

          <template v-else-if="column.key === 'publishUser'">
            <span>{{ record.createdStaffName || '-' }}</span>
          </template>

          <template v-else-if="column.key === 'publishTime'">
            <div class="homework-datetime-cell">
              <template v-if="hasDateTimeValue(record.publishTime)">
                <div>{{ formatDateText(record.publishTime) }}</div>
                <div
                  class="homework-datetime-cell__time"
                  :class="{ 'homework-datetime-cell__time--future': isFutureDateTime(record.publishTime) }"
                >
                  {{ isFutureDateTime(record.publishTime) ? `将于${formatTimeText(record.publishTime)}发布` : formatTimeText(record.publishTime) }}
                </div>
              </template>
              <span v-else>-</span>
            </div>
          </template>

          <template v-else-if="column.key === 'deadlineTime'">
            <div class="homework-datetime-cell">
              <template v-if="hasDateTimeValue(record.endTime)">
                <div>{{ formatDateText(record.endTime) }}</div>
                <div class="homework-datetime-cell__time">
                  {{ formatTimeText(record.endTime) }}
                </div>
              </template>
              <span v-else>-</span>
            </div>
          </template>

          <template v-else-if="column.key === 'action'">
            <a-space :size="12">
              <a class="font-500" @click="openEditModal(record)">编辑</a>
              <a class="font-500 text-#ff4d4f" @click="handleDelete(record)">删除</a>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>

    <afterSchoolTasksModel
      v-model="modalOpen"
      :mode="modalMode"
      :homework-id="currentHomeworkId"
      @success="handleModalSuccess"
    />
  </div>
</template>

<style scoped lang="less">
.homework-page {
  .homework-panel__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .homework-panel__summary {
    display: flex;
    align-items: center;
    min-width: 0;
  }

  .homework-panel__title {
    position: relative;
    padding-left: 10px;
    color: #262626;
    font-size: 16px;
    font-weight: 600;
    line-height: 24px;

    &::before {
      position: absolute;
      top: 6px;
      left: 0;
      width: 4px;
      height: 12px;
      border-radius: 999px;
      background: var(--pro-ant-color-primary);
      content: '';
    }
  }

  .homework-name-cell__title {
    color: #262626;
    font-size: 14px;
    line-height: 22px;
    font-weight: 500;
  }

  .homework-name-cell__meta {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 6px;
    color: #8c8c8c;
    font-size: 12px;
    line-height: 20px;
  }

  .homework-source-tag {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 44px;
    height: 22px;
    padding: 0 8px;
    border-radius: 999px;
    background: #edf5ff;
    color: var(--pro-ant-color-primary);
    font-size: 12px;
    line-height: 20px;

    &.is-one-to-one {
      background: #fff4eb;
      color: #fa8c16;
    }
  }

  .homework-name-cell__source {
    min-width: 0;
    color: #8c8c8c;
  }

  .homework-content-cell {
    max-width: 220px;
    color: #262626;
    line-height: 22px;
  }

  .homework-metric-cell__value {
    color: #262626;
    font-size: 14px;
    line-height: 22px;
    font-weight: 500;
  }

  .homework-metric-cell__desc,
  .homework-datetime-cell__time {
    color: #8c8c8c;
    font-size: 12px;
    line-height: 20px;
  }

  .homework-datetime-cell__time--future {
    color: #fa8c16;
    font-weight: 500;
  }

  .homework-column-title {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .homework-column-title__icon {
    color: #999;
    font-size: 14px;
    cursor: pointer;
    transition: color 0.2s ease;

    &:hover {
      color: var(--pro-ant-color-primary);
    }
  }
}
</style>
