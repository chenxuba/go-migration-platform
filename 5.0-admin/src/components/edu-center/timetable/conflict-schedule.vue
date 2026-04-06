<script setup lang="ts">
import dayjs from 'dayjs'
import { Modal } from 'ant-design-vue'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import type { TeachingScheduleItem } from '@/api/edu-center/teaching-schedule'
import { cancelTeachingSchedulesApi, listTeachingSchedulesApi } from '@/api/edu-center/teaching-schedule'
import emitter, { EVENTS } from '@/utils/eventBus'
import messageService from '@/utils/messageService'

type ConflictFilterKey = 'teacher' | 'classroom' | 'student' | 'assistant'
type ConflictType = '老师' | '教室' | '学员' | '助教'

const weekdayLabels = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']

const displayArray = ['createTime', 'scheduleType']

const conflictFilterTypeMap: Record<ConflictFilterKey, ConflictType> = {
  teacher: '老师',
  classroom: '教室',
  student: '学员',
  assistant: '助教',
}

const scheduleTypeOptions = [
  { id: 'teacher', value: '上课教师冲突' },
  { id: 'classroom', value: '上课教室冲突' },
  { id: 'student', value: '学员冲突' },
  { id: 'assistant', value: '助教冲突' },
]

const tableColumns = [
  {
    title: '上课日期/时段',
    key: 'dateTime',
    width: 220,
  },
  {
    title: '日程类型',
    key: 'scheduleType',
    width: 130,
  },
  {
    title: '所属班级/1对1',
    key: 'teachingClass',
    width: 210,
  },
  {
    title: '所属课程',
    key: 'lessonName',
    width: 130,
  },
  {
    title: '上课教师',
    key: 'teacher',
    width: 120,
  },
  {
    title: '教室',
    key: 'classroom',
    width: 100,
  },
  {
    title: '操作',
    key: 'actions',
    width: 130,
  },
]

function getWeekRange(base = dayjs()) {
  const current = dayjs(base)
  const diff = current.day() === 0 ? -6 : 1 - current.day()
  const start = current.add(diff, 'day').startOf('day')
  const end = start.add(6, 'day')
  return [
    start.format('YYYY-MM-DD'),
    end.format('YYYY-MM-DD'),
  ]
}

function normalizeFilterValues(value: unknown) {
  if (!Array.isArray(value))
    return []
  return value.map(item => String(item ?? '').trim()).filter(Boolean)
}

const defaultCreateTimeVals = getWeekRange()
const queryRange = ref<string[]>([...defaultCreateTimeVals])
const selectedConflictKeys = ref<ConflictFilterKey[]>([])
const loading = ref(false)
const deletingId = ref('')
const scheduleRows = ref<TeachingScheduleItem[]>([])
let requestToken = 0

const selectedConflictTypes = computed(() =>
  selectedConflictKeys.value
    .map(key => conflictFilterTypeMap[key])
    .filter(Boolean),
)

const conflictRows = computed(() => {
  const filtered = scheduleRows.value.filter((item) => {
    if (!item?.conflict)
      return false
    if (!selectedConflictTypes.value.length)
      return true
    return selectedConflictTypes.value.some(type => (item.conflictTypes || []).includes(type))
  })

  return filtered.slice().sort((a, b) => {
    const timeDiff = dayjs(a.startAt).valueOf() - dayjs(b.startAt).valueOf()
    if (timeDiff !== 0)
      return timeDiff
    return String(a.id || '').localeCompare(String(b.id || ''))
  })
})

function handleCreateTimeFilter(value: unknown) {
  if (Array.isArray(value) && value.length === 2) {
    queryRange.value = [
      String(value[0] || ''),
      String(value[1] || ''),
    ]
    return
  }
  queryRange.value = [...defaultCreateTimeVals]
}

function handleConflictTypeFilter(value: unknown) {
  selectedConflictKeys.value = normalizeFilterValues(value) as ConflictFilterKey[]
}

async function loadConflictSchedules() {
  const [startDate, endDate] = queryRange.value
  const token = ++requestToken
  loading.value = true
  try {
    const res = await listTeachingSchedulesApi({
      startDate,
      endDate,
    })
    if (token !== requestToken)
      return
    scheduleRows.value = res.code === 200 && Array.isArray(res.result) ? res.result : []
  }
  catch (error) {
    if (token !== requestToken)
      return
    console.error('load conflict schedules failed', error)
    scheduleRows.value = []
    messageService.error(error?.response?.data?.message || error?.message || '加载冲突日程失败')
  }
  finally {
    if (token === requestToken)
      loading.value = false
  }
}

watch(
  queryRange,
  () => {
    loadConflictSchedules()
  },
  { deep: true, immediate: true },
)

function handleExternalRefresh() {
  loadConflictSchedules()
}

onMounted(() => {
  emitter.on(EVENTS.REFRESH_DATA, handleExternalRefresh)
})

onUnmounted(() => {
  emitter.off(EVENTS.REFRESH_DATA, handleExternalRefresh)
})

function formatLessonDate(record: Record<string, any>) {
  const value = dayjs(record.lessonDate)
  return `${value.format('YYYY-MM-DD')}（${weekdayLabels[value.day()] || ''}）`
}

function formatLessonTime(record: Record<string, any>) {
  return `${dayjs(record.startAt).format('HH:mm')} ~ ${dayjs(record.endAt).format('HH:mm')}`
}

function scheduleTypeLabel(record: Record<string, any>) {
  return Number(record.classType) === 2 ? '1对1日程' : '班课日程'
}

function teachingClassBadge(record: Record<string, any>) {
  return Number(record.classType) === 2 ? '1v1' : '班课'
}

function teachingClassBadgeClass(record: Record<string, any>) {
  return Number(record.classType) === 2
    ? 'text-#ff7d7d bg-#fff1f1 rounded-2.5 inline-block text-2.5 pt-0.5 pb-0.5 pl-1.5 pr-1.5'
    : 'text-#d46b08 bg-#fff5e8 rounded-2.5 inline-block text-2.5 pt-0.5 pb-0.5 pl-1.5 pr-1.5'
}

function isOneToOneSchedule(record: Record<string, any>) {
  return Number(record.classType) === 2
}

function buildDeleteSuccessMessage(record: Record<string, any>) {
  const dateText = dayjs(record.lessonDate).format('M月D日')
  return `已删除 ${dateText} ${formatLessonTime(record)} 的 1对1 冲突日程`
}

function confirmDelete(record: Record<string, any>) {
  if (!isOneToOneSchedule(record)) {
    messageService.info('班课删除暂未开放，后续再补主教/辅教联动删除。')
    return
  }

  Modal.confirm({
    centered: true,
    title: '确认删除这条冲突日程？',
    content: '删除后会同步从当前 1 对 1 的排课记录中移除，该操作不可恢复。',
    okText: '删除',
    cancelText: '取消',
    okButtonProps: {
      danger: true,
    },
    onOk: async () => {
      deletingId.value = String(record.id || '')
      try {
        const res = await cancelTeachingSchedulesApi({
          ids: [String(record.id)],
        })
        if (res.code !== 200)
          throw new Error(res.message || '删除冲突日程失败')
        messageService.success(buildDeleteSuccessMessage(record))
        emitter.emit(EVENTS.REFRESH_DATA)
      }
      catch (error) {
        console.error('delete conflict schedule failed', error)
        messageService.error(error?.response?.data?.message || error?.message || '删除冲突日程失败')
        throw error
      }
      finally {
        deletingId.value = ''
      }
    },
  })
}
</script>

<template>
  <div class="conflict-page">
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0">
      <all-filter
        type="noDelCreateTime"
        :display-array="displayArray"
        :default-create-time-vals="defaultCreateTimeVals"
        create-time-label="上课日期"
        :schedule-type-options="scheduleTypeOptions"
        schedule-type-label="冲突类型"
        @update:create-time-filter="handleCreateTimeFilter"
        @update:schedule-type-filter="handleConflictTypeFilter"
      />
    </div>

    <div class="conflict-board mt-2 bg-white rounded-4">
      <div class="conflict-board__head">
        <div class="conflict-board__title">
          共 {{ conflictRows.length }} 条冲突日程
        </div>
      </div>

      <a-table
        row-key="id"
        size="small"
        :columns="tableColumns"
        :data-source="conflictRows"
        :pagination="false"
        :loading="loading"
        :scroll="{ x: 980 }"
        :locale="{ emptyText: '当前筛选条件下暂无冲突日程' }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'dateTime'">
            <div class="flex flex-col">
              <div class="text-3.5 text-#333 whitespace-nowrap">
                {{ formatLessonDate(record) }}
              </div>
              <div class="text-3 text-#888 whitespace-nowrap">
                {{ formatLessonTime(record) }}
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'scheduleType'">
            <div class="text-#222 whitespace-nowrap">
              {{ scheduleTypeLabel(record) }}
            </div>
          </template>

          <template v-else-if="column.key === 'teachingClass'">
            <div class="flex items-center gap-1.5">
              <span :class="teachingClassBadgeClass(record)">
                {{ teachingClassBadge(record) }}
              </span>
              <span class="text-#ff6b6b whitespace-nowrap">
                {{ record.teachingClassName || '-' }}
              </span>
            </div>
          </template>

          <template v-else-if="column.key === 'lessonName'">
            <div class="text-#222 whitespace-nowrap">
              {{ record.lessonName || '-' }}
            </div>
          </template>

          <template v-else-if="column.key === 'teacher'">
            <div class="text-#222 whitespace-nowrap">
              {{ record.teacherName || '-' }}
            </div>
          </template>

          <template v-else-if="column.key === 'classroom'">
            <div class="text-#222 whitespace-nowrap">
              {{ record.classroomName || '-' }}
            </div>
          </template>

          <template v-else-if="column.key === 'actions'">
            <a-space :size="8">
              <a-tooltip title="编辑暂未开放">
                <span>
                  <a-button type="link" size="small" disabled>
                    编辑
                  </a-button>
                </span>
              </a-tooltip>

              <a-tooltip v-if="!isOneToOneSchedule(record)" title="班课删除暂未开放">
                <span>
                  <a-button type="link" size="small" danger disabled>
                    删除
                  </a-button>
                </span>
              </a-tooltip>
              <a-button
                v-else
                type="link"
                size="small"
                danger
                :loading="deletingId === String(record.id)"
                @click="confirmDelete(record)"
              >
                删除
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>
  </div>
</template>

<style scoped lang="less">
.conflict-page {
  padding-top: 0;
}

.conflict-board {
  padding: 16px 18px 18px;
  box-shadow: 0 10px 28px rgba(15, 35, 95, 0.04);
}

.conflict-board__head {
  margin-bottom: 16px;
}

.conflict-board__title {
  position: relative;
  padding-left: 14px;
  color: #222;
  font-size: 14px;
}

.conflict-board__title::before {
  position: absolute;
  left: 0;
  top: 5px;
  width: 5px;
  height: 14px;
  border-radius: 999px;
  background: #1677ff;
  content: '';
}

</style>
