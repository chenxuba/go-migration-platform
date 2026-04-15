<script setup lang="ts">
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs, { type Dayjs } from 'dayjs'
import type { TableColumnsType } from 'ant-design-vue'
import { computed, h, ref, watch } from 'vue'
import { cancelTeachingScheduleScopedApi } from '@/api/edu-center/teaching-schedule'
import { getGroupClassDrawerWaitingRollCallSchedulesApi, type GroupClassDrawerWaitingRollCallScheduleItem } from '@/api/edu-center/group-class'
import RollCallDrawer from '@/components/common/roll-call-drawer.vue'
import GroupClassUnscheduledRollCallModal from '@/components/edu-center/class-list/group-class-unscheduled-roll-call-modal.vue'
import SmartTimetableScheduleDetailDrawer from '@/components/edu-center/timetable/smart-timetable-schedule-detail-drawer.vue'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  open?: boolean
  classId?: string
  className?: string
  studentCount?: number
}>(), {
  open: false,
  classId: '',
  className: '',
  studentCount: 0,
})

const displayArray = ref(['scheduleDate'])
const columns: TableColumnsType<GroupClassDrawerWaitingRollCallScheduleItem> = [
  {
    title: '上课日期/时段',
    dataIndex: 'lessonDate',
    key: 'lessonDate',
    width: 180,
  },
  {
    title: '课程名称',
    dataIndex: 'lessonName',
    key: 'lessonName',
    width: 180,
  },
  {
    title: '上课教师',
    dataIndex: 'teacherName',
    key: 'teacherName',
    width: 110,
  },
  {
    title: '上课助教',
    dataIndex: 'assistantText',
    key: 'assistantText',
    width: 120,
  },
  {
    title: '上课教室',
    dataIndex: 'classroomName',
    key: 'classroomName',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 150,
    fixed: 'right',
  },
]

const loading = ref(false)
const deleting = ref(false)
const rawList = ref<GroupClassDrawerWaitingRollCallScheduleItem[]>([])
const dateRange = ref<[Dayjs, Dayjs] | null>(null)
const allFilterKey = ref(0)
const defaultScheduleDateVals = ref<string[]>([])
const rollCallOpen = ref(false)
const unscheduledRollCallOpen = ref(false)
const currentScheduleId = ref('')
const currentLessonDay = ref('')
const detailOpen = ref(false)
const detailState = ref<Record<string, any> | null>(null)
let loadSeq = 0

function todayRange() {
  const today = dayjs().startOf('day')
  return [today, today.endOf('day')] as [Dayjs, Dayjs]
}

function todayRangeValues() {
  const today = dayjs().format('YYYY-MM-DD')
  return [today, today]
}

function isSameDateRange(nextRange: [Dayjs, Dayjs] | null) {
  if (!dateRange.value && !nextRange)
    return true
  if (!dateRange.value || !nextRange)
    return false
  return dateRange.value[0].isSame(nextRange[0], 'day') && dateRange.value[1].isSame(nextRange[1], 'day')
}

const totalWidth = computed(() =>
  columns.reduce((sum, item) => sum + Number(item.width || 0), 0),
)
const unscheduledRollCallDisabledReason = computed(() => {
  if (!String(props.classId || '').trim())
    return '当前班级信息不完整，暂不可创建未排课点名'
  if (Number(props.studentCount || 0) <= 0)
    return `${String(props.className || '当前班级').trim() || '当前班级'}暂无在班学员，不能创建未排课点名`
  return ''
})
const canCreateUnscheduledRollCall = computed(() => !unscheduledRollCallDisabledReason.value)

async function loadList() {
  const classId = String(props.classId || '').trim()
  if (!props.open || !classId) {
    rawList.value = []
    return
  }
  const currentSeq = ++loadSeq
  loading.value = true
  try {
    const res = await getGroupClassDrawerWaitingRollCallSchedulesApi({
      classId,
      startDate: dateRange.value?.[0]?.format('YYYY-MM-DD'),
      endDate: dateRange.value?.[1]?.format('YYYY-MM-DD'),
    })
    if (currentSeq !== loadSeq)
      return
    if (res.code !== 200)
      throw new Error(res.message || '加载待点名日程失败')
    rawList.value = Array.isArray(res.result?.list) ? res.result.list : []
  }
  catch (error: any) {
    if (currentSeq !== loadSeq)
      return
    rawList.value = []
    messageService.error(error?.response?.data?.message || error?.message || '加载待点名日程失败')
  }
  finally {
    if (currentSeq === loadSeq)
      loading.value = false
  }
}

function handleScheduleDateFilter(value: unknown) {
  if (!Array.isArray(value) || value.length < 2) {
    if (isSameDateRange(null))
      return
    dateRange.value = null
    if (props.open && String(props.classId || '').trim())
      loadList()
    return
  }
  const start = dayjs(String(value[0] || ''))
  const end = dayjs(String(value[1] || ''))
  if (!start.isValid() || !end.isValid()) {
    if (isSameDateRange(null))
      return
    dateRange.value = null
    if (props.open && String(props.classId || '').trim())
      loadList()
    return
  }
  const nextRange: [Dayjs, Dayjs] = [start.startOf('day'), end.endOf('day')]
  if (isSameDateRange(nextRange))
    return
  dateRange.value = nextRange
  if (props.open && String(props.classId || '').trim())
    loadList()
}

function handleRollCall(record: GroupClassDrawerWaitingRollCallScheduleItem | Record<string, any>) {
  if (!record.canRollCall) {
    messageService.info(record.rollCallDisabledReason || '当前日程暂不可点名')
    return
  }
  currentScheduleId.value = String(record.id || '').trim()
  currentLessonDay.value = String(record.lessonDate || '').trim()
  if (!currentScheduleId.value)
    return
  rollCallOpen.value = true
}

function handleViewDetail(record: GroupClassDrawerWaitingRollCallScheduleItem | Record<string, any>) {
  detailState.value = {
    scheduleId: record.id,
    id: record.id,
    batchNo: record.batchNo,
    batchSize: record.batchSize,
    lessonTitle: record.lessonName || props.className || '日程详情',
    assistantText: record.assistantText,
    classroomName: record.classroomName,
    teacherName: record.teacherName,
  }
  detailOpen.value = true
}

function handleCreateUnscheduledRollCall() {
  if (unscheduledRollCallDisabledReason.value) {
    messageService.warning(unscheduledRollCallDisabledReason.value)
    return
  }
  unscheduledRollCallOpen.value = true
}

function handleDelete(record: GroupClassDrawerWaitingRollCallScheduleItem | Record<string, any>) {
  const scheduleId = String(record.id || '').trim()
  if (!scheduleId || deleting.value)
    return
  Modal.confirm({
    title: '删除待点名日程',
    icon: h(ExclamationCircleOutlined),
    content: '删除后不可恢复，确定删除这条待点名日程吗？',
    async onOk() {
      deleting.value = true
      try {
        const res = await cancelTeachingScheduleScopedApi({
          id: scheduleId,
          scope: 'current',
        })
        if (res.code !== 200)
          throw new Error(res.message || '删除待点名日程失败')
        messageService.success('删除成功')
        await loadList()
      }
      catch (error: any) {
        messageService.error(error?.response?.data?.message || error?.message || '删除待点名日程失败')
        throw error
      }
      finally {
        deleting.value = false
      }
    },
  })
}

watch(
  () => `${props.open}|${String(props.classId || '').trim()}`,
  () => {
    dateRange.value = todayRange()
    defaultScheduleDateVals.value = todayRangeValues()
    allFilterKey.value += 1
    loadList()
  },
  { immediate: true },
)
</script>

<template>
  <div class="m-12px">
    <all-filter
      :key="allFilterKey"
      :display-array="displayArray"
      :default-schedule-date-vals="defaultScheduleDateVals"
      @update:schedule-date-filter="handleScheduleDateFilter"
    />
  </div>
  <div class="m-12px">
    <div class="bg-#fff pt-18px px-20px rounded-10px">
      <div class="flex justify-between items-center">
        <custom-title :title="`当前共计 ${rawList.length} 条待点名日程`" font-size="14px" class="pb-12px" />
        <a-tooltip :title="unscheduledRollCallDisabledReason || undefined">
          <span>
            <a-button type="primary" class="mb-12px" :disabled="!canCreateUnscheduledRollCall" @click="handleCreateUnscheduledRollCall">
              创建未排课点名
            </a-button>
          </span>
        </a-tooltip>
      </div>
      <a-table
        row-key="id"
        size="small"
        :loading="loading"
        :columns="columns"
        :data-source="rawList"
        :pagination="false"
        :scroll="{ x: totalWidth }"
      >
        <template #headerCell="{ column }">
          <template v-if="column.key === 'lessonDate'">
            上课日期/时段
            <a-popover title="上课日期/时段">
              <template #content>
                <div>展示当前班级尚未完成点名的有效日程</div>
              </template>
              <ExclamationCircleOutlined />
            </a-popover>
          </template>
        </template>
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'lessonDate'">
            <div>{{ dayjs(record.lessonDate).isValid() ? `${dayjs(record.lessonDate).format('YYYY-MM-DD')}(${['周日','周一','周二','周三','周四','周五','周六'][dayjs(record.lessonDate).day()]})` : '-' }}</div>
            <div>{{ `${dayjs(record.startAt).isValid() ? dayjs(record.startAt).format('HH:mm') : '--:--'} ～ ${dayjs(record.endAt).isValid() ? dayjs(record.endAt).format('HH:mm') : '--:--'}` }}</div>
          </template>
          <template v-if="column.dataIndex === 'lessonName'">
            {{ record.lessonName || '-' }}
          </template>
          <template v-if="column.dataIndex === 'teacherName'">
            {{ record.teacherName || '-' }}
          </template>
          <template v-if="column.dataIndex === 'assistantText'">
            {{ record.assistantText || '-' }}
          </template>
          <template v-if="column.dataIndex === 'classroomName'">
            {{ record.classroomName || '-' }}
          </template>
          <template v-if="column.dataIndex === 'action'">
            <a-space :size="12">
              <a @click="handleRollCall(record)">点名</a>
              <a @click="handleViewDetail(record)">详情</a>
              <a @click="handleDelete(record)">删除</a>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>
    <RollCallDrawer
      v-model:open="rollCallOpen"
      :schedule-id="currentScheduleId"
      :lesson-day="currentLessonDay"
      @updated="loadList"
      @confirmed="loadList"
    />
    <SmartTimetableScheduleDetailDrawer
      v-model:open="detailOpen"
      :detail="detailState"
      @updated="loadList"
    />
    <GroupClassUnscheduledRollCallModal
      v-model:open="unscheduledRollCallOpen"
      :class-id="props.classId"
      @updated="loadList"
      @confirmed="loadList"
    />
  </div>
</template>
